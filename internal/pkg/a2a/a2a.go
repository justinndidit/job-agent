package a2a

type A2AError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type MessagePart struct {
	Kind     string                 `json:"kind"`
	Text     string                 `json:"text,omitempty"`
	Data     interface{}            `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type A2AMessage struct {
	Role             string                 `json:"role"`
	Parts            []MessagePart          `json:"parts"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Extensions       []string               `json:"extensions,omitempty"`
	ReferenceTaskIds []string               `json:"referenceTaskIds,omitempty"`
	MessageID        string                 `json:"messageId"`
	TaskID           string                 `json:"taskId,omitempty"`
	ContextID        string                 `json:"contextId,omitempty"`
	Kind             string                 `json:"kind"`
}

type PushNotificationConfig struct {
	Url            string `json:"url"`
	Token          string `json:"token omitempty"`
	Authentication string `json:"authentication omitempty"`
}

type MessageConfiguration struct {
	Blocking               string                 `json:"blocking"`
	AcceptedOutputModes    string                 `json:"acceptedOutputModes"`
	PushNotificationConfig PushNotificationConfig `json:"pushNotificationConfig omitempty"`
}

type MessageParams struct {
	Message       A2AMessage           `json:"message"`
	Configuration MessageConfiguration `json:"configuration"`
}

type ExecuteParams struct {
	ContextID string       `json:"contextId omitempty"`
	TaskID    string       `json:"taskId omitempty"`
	Messages  []A2AMessage `json:"messages"`
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      string      `json:"id"`
}

type Task struct {
	ID        string                 `json:"id"`
	State     string                 `json:"state"`
	Message   *A2AMessage            `json:"message,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type Artifact struct {
	ArtifactID string        `json:"artifactId"`
	Name       string        `json:"name"`
	Parts      []MessagePart `json:"parts"`
}

type TaskResult struct {
	ID        string       `json:"id"`
	ContextID string       `json:"contextId"`
	Status    Task         `json:"status"`
	Artifacts []Artifact   `json:"artifacts"`
	History   []A2AMessage `json:"history"`
	Kind      string       `json:"kind"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *A2AError   `json:"error,omitempty"`
}

// Agent Card (for A2A discovery)
type AgentCard struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Version      string            `json:"version"`
	Capabilities []string          `json:"capabilities"`
	Skills       []Skill           `json:"skills"`
	ServiceURL   string            `json:"serviceUrl"`
	Auth         map[string]string `json:"auth,omitempty"`
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
