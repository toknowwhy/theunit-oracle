package testutil

import (
	"errors"
	"time"

	"github.com/toknowwhy/theunit-oracle/pkg/gofer"
	"github.com/toknowwhy/theunit-oracle/pkg/gofer/graph"
	"github.com/toknowwhy/theunit-oracle/pkg/gofer/graph/nodes"
)

func Gofer(ps ...gofer.Pair) gofer.Gofer {
	graphs := map[gofer.Pair]nodes.Aggregator{}
	for _, p := range ps {
		root := nodes.NewMedianAggregatorNode(p, 1)

		ttl := time.Second * time.Duration(time.Now().Unix()+10)
		on1 := nodes.NewOriginNode(nodes.OriginPair{Origin: "a", Pair: p}, 0, ttl)
		on2 := nodes.NewOriginNode(nodes.OriginPair{Origin: "b", Pair: p}, 0, ttl)
		in := nodes.NewIndirectAggregatorNode(p)
		mn := nodes.NewMedianAggregatorNode(p, 1)

		root.AddChild(on1)
		root.AddChild(in)
		root.AddChild(mn)

		in.AddChild(on1)
		mn.AddChild(on1)
		mn.AddChild(on2)

		_ = on1.Ingest(nodes.OriginPrice{
			PairPrice: nodes.PairPrice{
				Pair:      p,
				Price:     10,
				Bid:       10,
				Ask:       10,
				Volume24h: 10,
				Time:      time.Unix(10, 0),
			},
			Origin: "a",
			Error:  nil,
		})

		_ = on2.Ingest(nodes.OriginPrice{
			PairPrice: nodes.PairPrice{
				Pair:      p,
				Price:     20,
				Bid:       20,
				Ask:       20,
				Volume24h: 20,
				Time:      time.Unix(20, 0),
			},
			Origin: "b",
			Error:  errors.New("something"),
		})

		graphs[p] = root
	}

	return graph.NewGofer(graphs, nil)
}

func Models(ps ...gofer.Pair) map[gofer.Pair]*gofer.Model {
	g := Gofer(ps...)
	ns, err := g.Models()
	if err != nil {
		panic(err)
	}
	return ns
}

func Prices(ps ...gofer.Pair) map[gofer.Pair]*gofer.Price {
	g := Gofer(ps...)
	ts, err := g.Prices()
	if err != nil {
		panic(err)
	}
	return ts
}
