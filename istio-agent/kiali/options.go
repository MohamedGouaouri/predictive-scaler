package kiali

type WorkloadGraph struct {
	Timestamp int64    `json:"timestamp"`
	Duration  int      `json:"duration"`
	GraphType string   `json:"graphType"`
	Elements  Elements `json:"elements"`
}

type Elements struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	Data NodeData `json:"data"`
}

type NodeData struct {
	ID           string        `json:"id"`
	NodeType     string        `json:"nodeType"`
	Cluster      string        `json:"cluster"`
	Namespace    string        `json:"namespace"`
	Workload     string        `json:"workload"`
	App          string        `json:"app"`
	Version      string        `json:"version"`
	DestServices []DestService `json:"destServices"`
	Traffic      []Traffic     `json:"traffic"`
	HealthData   HealthData    `json:"healthData"`
	IsGateway    *Gateway      `json:"isGateway,omitempty"`
	IsOutside    bool          `json:"isOutside,omitempty"`
	IsRoot       bool          `json:"isRoot,omitempty"`
	Cpu          float64       `json:"cpu,omitempty"`
	Mem          float64       `json:"memory,omitempty"`
	Replicas     int           `json:"replicas,omitempty"`
}

type DestService struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type Traffic struct {
	Protocol string            `json:"protocol"`
	Rates    map[string]string `json:"rates"`
}

type HealthData struct {
	WorkloadStatus    WorkloadStatus         `json:"workloadStatus"`
	Requests          Requests               `json:"requests"`
	HealthAnnotations map[string]interface{} `json:"healthAnnotations"`
}

type WorkloadStatus struct {
	Name              string `json:"name"`
	DesiredReplicas   int    `json:"desiredReplicas"`
	CurrentReplicas   int    `json:"currentReplicas"`
	AvailableReplicas int    `json:"availableReplicas"`
	SyncedProxies     int    `json:"syncedProxies"`
}

type Requests struct {
	Inbound  RequestDetails `json:"inbound"`
	Outbound RequestDetails `json:"outbound"`
}

type RequestDetails struct {
	HTTP map[string]float64 `json:"http"`
}

type Gateway struct {
	IngressInfo    IngressInfo    `json:"ingressInfo"`
	EgressInfo     EgressInfo     `json:"egressInfo"`
	GatewayAPIInfo GatewayAPIInfo `json:"gatewayAPIInfo"`
}

type IngressInfo struct {
	Hostnames []string `json:"hostnames"`
}

type EgressInfo struct{}

type GatewayAPIInfo struct{}

type Edge struct {
	Data EdgeData `json:"data"`
}

type EdgeData struct {
	ID           string      `json:"id"`
	Source       string      `json:"source"`
	Target       string      `json:"target"`
	ResponseTime string      `json:"responseTime"`
	Traffic      EdgeTraffic `json:"traffic"`
}

type EdgeTraffic struct {
	Protocol  string              `json:"protocol"`
	Rates     map[string]string   `json:"rates"`
	Responses map[string]Response `json:"responses"`
}

type Response struct {
	Flags map[string]string `json:"flags"`
	Hosts map[string]string `json:"hosts"`
}

type ReshapedNode struct {
	ID        string          `json:"id"`
	Workload  string          `json:"workload"`
	IsGateway bool            `json:"IsGateway,omitempty"`
	Cpu       float64         `json:"cpu,omitempty"`
	Mem       float64         `json:"memory,omitempty"`
	Replicas  int             `json:"replicas,omitempty"`
	Children  []*ReshapedNode `json:"children,omitempty"`
	Edges     []ReshapedEdge  `json:"edges,omitempty"`
}

type ReshapedEdge struct {
	ResponseTime float64 `json:"responseTime"`
	RequestRate  float64 `json:"requestRate"`
}
