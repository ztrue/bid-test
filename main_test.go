package main

import "encoding/json"
import "errors"
import "io/ioutil"
import "net/http"
import "testing"
import "time"

const RootURL = "http://localhost:8080/"

type Result struct {
  Price int `json:"price"`
  Source string `json:"source"`
}

type TestServiceCase struct {
  url string
  paths []string
  status int
  result Result
  err error
}

func TestService(t *testing.T) {
  m := NewSourceMock("localhost:9000")
  m.SetUp(
    "primes",
    http.StatusOK,
    jsonPrices(2, 3, 5, 7, 11, 13, 17, 19, 23),
  )
  m.SetUp(
    "fibo",
    http.StatusOK,
    jsonPrices(1, 1, 2, 3, 5, 8, 13, 21),
  )
  m.SetUp(
    "fact",
    http.StatusOK,
    jsonPrices(1, 2, 6, 24),
  )
  m.SetUp(
    "rand",
    http.StatusOK,
    jsonPrices(5, 17, 3, 19, 76, 24, 1, 5, 10, 34, 8, 27, 7),
  )
  m.SetUp(
    "low",
    http.StatusOK,
    jsonPrices(3, 2, 1),
  )
  m.SetUp(
    "single",
    http.StatusOK,
    jsonPrices(17),
  )
  m.SetUp(
    "empty",
    http.StatusOK,
    jsonPrices(),
  )
  m.SetUp(
    "unavailable",
    http.StatusServiceUnavailable,
    "",
  )
  m.SetUp(
    "timeout",
    http.StatusServiceUnavailable,
    jsonPrices(1, 42, 1337, 100500),
    time.Duration(120) * time.Millisecond,
  )

  // Run source service mock
  go m.Run()
  // Run tested service
  go main()

  cases := []TestServiceCase{
    // Competition
    {
      "",
      []string{"fibo", "primes"},
      http.StatusOK,
      Result{21, m.URL("primes")},
      nil,
    },
    {
      "",
      []string{"primes", "fibo", "fact", "low"},
      http.StatusOK,
      Result{23, m.URL("fact")},
      nil,
    },
    {
      "",
      []string{"primes", "rand", "low"},
      http.StatusOK,
      Result{34, m.URL("rand")},
      nil,
    },
    // Duplicated sources are OK
    {
      "",
      []string{"rand", "rand", "rand", "rand"},
      http.StatusOK,
      Result{76, m.URL("rand")},
      nil,
    },
    // Only one price available (OK and this price is counted as result)
    {
      "",
      []string{"single"},
      http.StatusOK,
      Result{17, m.URL("single")},
      nil,
    },
    // Only one source available with prices
    {
      "",
      []string{"timeout", "unavailable", "empty", "low"},
      http.StatusOK,
      Result{2, m.URL("low")},
      nil,
    },
    // No prices available
    {
      "",
      []string{"timeout", "unavailable", "empty"},
      http.StatusUnprocessableEntity,
      Result{},
      errors.New("There is no prices"),
    },
    // No sources passed
    {
      "",
      []string{},
      http.StatusBadRequest,
      Result{},
      errors.New("There must be at least one source"),
    },
    // Wrong URL
    {
      RootURL + "/loser",
      []string{"fact", "rand"},
      http.StatusNotFound,
      Result{},
      errors.New("Requested path not found"),
    },
  }

  for i, c := range cases {
    url := c.url
    if url == "" {
      sources := []string{}
      for _, path := range c.paths {
        sources = append(sources, m.URL(path))
      }
      url = getURL(sources)
    }

    status, result, err := doRequest(t, url)

    if err != nil && (c.err == nil || err.Error() != c.err.Error()) {
      t.Errorf("Case %d, want: `%s`, got: `%s`", i, c.err, err)
    }
    if status != c.status {
      t.Errorf("Case %d, want: `%d`, got: `%d`", i, c.status, status)
    }
    if result != c.result {
      t.Errorf("Case %d, want: `%v`, got: `%v`", i, c.result, result)
    }
  }
}

func jsonPrices(prices ...int) string {
  data := []map[string]int{}
  for _, p := range prices {
    data = append(data, map[string]int{"price": p})
  }
  body, _ := json.Marshal(data)
  return string(body)
}


func getURL(sources []string) string {
  url := RootURL + "/winner?"
  for i := 0; i < len(sources); i++ {
    if i > 0 {
      url += "&"
    }
    url += "s=" + sources[i]
  }
  return url
}

func doRequest(t *testing.T, url string) (int, Result, error) {
  res, err := http.Get(url)
  if err != nil {
    t.Errorf(err.Error())
  }
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    t.Errorf(err.Error())
  }
  if res.StatusCode != http.StatusOK {
    data := map[string]string{}
    if err := json.Unmarshal(body, &data); err != nil {
      t.Errorf("%v", data)
    }
    return res.StatusCode, Result{}, errors.New(data["error"])
  }
  data := Result{}
  if err := json.Unmarshal(body, &data); err != nil {
    t.Errorf("%v", data)
  }
  return res.StatusCode, data, nil
}
