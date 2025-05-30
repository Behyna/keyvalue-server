package keyvalue

const (
	CONNECTIONTYPE = "tcp"
)

// commands.
const (
	PutCommand    = "Put"
	GetCommand    = "Get"
	DeleteCommand = "Delete"
	UpdateCommand = "Update"
)

const CommandSeparator = ":"

const (
	clientConnect clientEventType = iota
	clientDisconnect
)
