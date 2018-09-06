package bid

import "encoding/json"
import "io/ioutil"
import "net/http"
import "sync"

// SourceData is data that source returns in response
type SourceData []SourceDataItem

// SourceDataItem is a single price item from source response
type SourceDataItem struct {
  Price int `json:"price"`
}

// SourceDetails is combined data about source and it's prices
type SourceDetails struct {
  Source string
  SourceData SourceData
}

// Result is a winner of bid round that contains source and second price
type Result struct {
  Price int `json:"price"`
  Source string `json:"source"`
}

// IsValid returns if result is valid
// Result counts as not valid if there is no source or there is no price
func (r Result) IsValid() bool {
  return r.Price > 0 && r.Source != ""
}

// Bid requests data from sources and returns the winner
func Bid(client *http.Client, sources []string) Result {
  return Calculate(getSourceDetails(client, sources))
}

// Calculate calculates the winner based on sources and prices
func Calculate(details chan SourceDetails) Result {
  var first Result
  var second Result
  for d := range details {
    for _, item := range d.SourceData {
      if first.Price == 0 || item.Price > first.Price {
        second = first
        first = Result{
          Price: item.Price,
          Source: d.Source,
        }
      } else if second.Price == 0 || item.Price > second.Price {
        second = Result{
          Price: item.Price,
          Source: d.Source,
        }
      }
    }
  }
  // First price counted as OK if there is no second price
  if second.Price > 0 {
    first.Price = second.Price
  }
  return first
}

func getSourceDetails(client *http.Client, sources []string) chan SourceDetails {
  wg := &sync.WaitGroup{}
  wg.Add(len(sources))
  ch := make(chan SourceDetails, len(sources))
  // TODO There could be a pool of workers, it can increase performance
  for _, source := range sources {
    go func (source string) {
      defer wg.Done()
      requestPrices(client, source, ch)
    }(source)
  }
  wg.Wait()
  close(ch)
  return ch
}

func requestPrices(client *http.Client, source string, ch chan SourceDetails) {
  res, err := client.Get(source)
  if err != nil {
    // TODO Log error if needed
    return
  }
  if res.StatusCode != http.StatusOK {
    // TODO Log error if needed
    return
  }
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    // TODO Log error if needed
    return
  }
  defer res.Body.Close()
  data := SourceData{}
  if err := json.Unmarshal(body, &data); err != nil {
    // TODO Log error if needed
    return
  }
  ch <- SourceDetails{
    Source: source,
    SourceData: data,
  }
}
