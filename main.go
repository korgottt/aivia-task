//nolint:gomnd
package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aiviaio/go-binance/v2"
)

type Client struct {
	bClient  *binance.Client
	wg       *sync.WaitGroup
	resultCh chan map[string]string
}

func main() {
	ctx := context.Background()

	c := &Client{bClient: binance.NewClient("", ""), wg: &sync.WaitGroup{}, resultCh: make(chan map[string]string, 5)}

	symbols := c.getSymbols(ctx)
	c.fillSymbolPriceChan(ctx, symbols)
	c.closeChan()
	c.displayResult()
}

func (c *Client) getSymbols(ctx context.Context) []string {
	resp, err := c.bClient.NewExchangeInfoService().Do(ctx)
	if err != nil {
		os.Exit(1)
	}

	symbols := make([]string, 5)

	for i := 0; i < len(symbols); i++ {
		symbols[i] = resp.Symbols[i].Symbol
	}

	return symbols
}

func (c *Client) fillSymbolPriceChan(
	ctx context.Context,
	symbols []string) {
	for _, s := range symbols {
		c.wg.Add(1)

		s := s

		go func() {
			defer c.wg.Done()

			resp, err := c.bClient.NewListPricesService().Symbol(s).Do(ctx)
			if err != nil || len(resp) == 0 {
				return
			}

			select {
			case c.resultCh <- map[string]string{resp[0].Symbol: resp[0].Price}:
			case <-ctx.Done():
				return
			}
		}()
	}
}

func (c *Client) closeChan() {
	go func() {
		c.wg.Wait()
		close(c.resultCh)
	}()
}

func (c *Client) displayResult() {
	for m := range c.resultCh {
		for k, v := range m {
			fmt.Printf("%s:%s\n", k, v)
		}
	}
}
