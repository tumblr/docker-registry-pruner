package graph

import (
	g "gonum.org/v1/gonum/graph"
)

type TagNode struct {
	Tag    string
	Layers []LayerNode
}
