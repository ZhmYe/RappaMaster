package paradigm

import (
	"encoding/json"
)

type Graph struct {
	Directed   bool                   `json:"directed"`
	MultiGraph bool                   `json:"multigraph"`
	GraphData  map[string]interface{} `json:"graph,omitempty"` // 全局属性
	Nodes      []Node                 `json:"nodes"`           // 节点列表
	Links      []Link                 `json:"links"`           // 边列表
}

type Node struct {
	ID    int                    // 固定字段
	Attrs map[string]interface{} // 动态属性
}

func (n Node) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{"id": n.ID}
	for k, v := range n.Attrs {
		out[k] = v
	}
	return json.Marshal(out)
}

func (n *Node) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	n.ID = int(raw["id"].(float64))
	delete(raw, "id")
	n.Attrs = raw
	return nil
}

type Link struct {
	Source int                    // 固定字段
	Target int                    // 固定字段
	Attrs  map[string]interface{} // 动态属性
}

func (l Link) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"source": l.Source,
		"target": l.Target,
	}
	for k, v := range l.Attrs {
		out[k] = v
	}
	return json.Marshal(out)
}

func (l *Link) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	l.Source = int(raw["source"].(float64))
	l.Target = int(raw["target"].(float64))
	delete(raw, "source")
	delete(raw, "target")
	l.Attrs = raw
	return nil
}
