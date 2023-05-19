//nolint:gomnd
package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aiviaio/go-binance/v2"
)

func main() {
	ctx := context.Background()
	bClient := binance.NewClient("", "")

	resp, err := bClient.NewExchangeInfoService().Do(ctx)
	if err != nil {
		os.Exit(1)
	}

	symbols := make([]string, 5)

	for i := 0; i < len(symbols); i++ {
		symbols[i] = resp.Symbols[i].Symbol
	}

	wg := &sync.WaitGroup{}
	resultCh := make(chan map[string]string, 5)

	for _, s := range symbols {
		wg.Add(1)

		s := s

		go func() {
			defer wg.Done()

			resp, err := bClient.NewListPricesService().Symbol(s).Do(ctx)
			if err != nil || len(resp) == 0 {
				return
			}

			select {
			case resultCh <- map[string]string{resp[0].Symbol: resp[0].Price}:
				return
			case <-ctx.Done():
				return
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for m := range resultCh {
		for k, v := range m {
			fmt.Printf("%s:%s\n", k, v)
		}
	}
}
