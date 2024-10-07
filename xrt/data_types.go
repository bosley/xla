package xrt

import "time"

const (
	ResourceTypeWeb byte = iota
	ResourceTypeFile
	ResourceTypeMessageChain
)

// File and resources used
type Resource struct {
	Id    string            // Unique Resource ID
	RType byte              // (ResourceType<TYPE> from const block above)
	Value string            // The value of the resouce (filepath, url, connection string, message id referencing Message struct, etc)
	Meta  map[string]string // Extension type, url type (https/ tcp/ etc)
}

// Authorization tokens users have to perfom actions (unique to user)
type TokenData struct {
	Id       string
	UserId   string // store this, not *User as it will be part of the encoded object
	Granted  time.Time
	Duration time.Duration
}

// Token value once "on the wire" with a "TTL" server-side the "sent" must authorize within
// the auth window, actual token info internal, nonce added, signature is passport signature
// from user of the Sent + Data + Nonce
type Token struct {
	Sent      time.Time // Must be within last minute (server restriction)
	Data      TokenData // Value of actual user token info
	Nonce     uint64    // Random value sent with
	Signature []byte    // Signature of the token data for verification
}

type Passport struct {
	Id    string // unique passport id
	Owner *User  // The owner of the passport
	// TODO: Essentially make a better "badger" badge with what we've learned from Badger
}

type User struct {
	Id           string       // unique uuidv4 to mark user
	Name         string       // Their display name (no spaces, must be unique)
	PasswordHash []byte       // bcrypt hash of user password
	Email        string       // Must be valid email in form, but we won't validate for now
	Tokens       []*TokenData // The data values of the tokens the user created for request auth
	Sessions     []*Session   // ids of all sessions user has going
	IsAgent      bool         // true iff its an agent
}

// Messages can contain message chains from multiple agents interacting
type Message struct {
	MessageId string            // The unique message id
	TopicId   string            // unique id to denote what the message chain is in regards to (which project)
	IdTo      string            // User Id (Agents are users too)
	IdFrom    string            // User Id (Agents are users too)
	Sent      time.Time         //  Time message left sender
	Received  time.Time         // Time message received by receiver
	Contents  string            // Contents of message
	Meta      map[string]string // Values like "role" "sentiment" "etc" are mapped here
	Previous  *Message          // The previous message (linked list)
	Next      *Message          // The next message
}

// Any interface to the system on behalf of "any" user (agents too)
// that is active (even if paused <on disk>)
type Session struct {
	Id             string       // UUIDv4
	UserId         *User        // references User struct id
	LocalResources []*Resource  // Resources being referenced in-session
	Chat           Conversation // The conversation occuring in the session
}

type Conversation struct {
	Id            string    // Unique id
	Started       time.Time // When chat started
	Participants  []*User   // Everyone involved
	RootMessageId *Message  // First message of message history
	Active        bool      // If the conversation has not yet been "finished"
}

type ModelSettings struct {
	Temperature    float32  // Default to: 0.7
	ModelNameTag   string   // Default to: llama3.1:latest
	SetupPrompt    string   // Specially set-aside setup prompt
	MessagePreLoad *Message // Messages that we can pre-load into the llm. Meta map will contain "role" for system/user distinction of message type
}

// A tool specially crafted for an ai agent to use to
// operate with web/fs/eetc
/*
	Since LLMS are probalistic, the Activation string will be used as "activation phrases"
	that can be used against an agen't response to see if the tool should be invoked.

	If invoked, the tool will can the message and attempt to figure out what they want

	If able to determine with %90 probablility, then the tool will trigger.
	The tool controller on the work bench will handle any "rollbacks" or "undos"
	from tools
*/
type SystemTool struct {
	Id          string   // unique id
	Description string   // what the tool does
	AgentPrompt string   // directon for general-purpose AI to attempt using the tool
	ResourceId  string   // the resource id of the item
	Activations []string // activators for tool (ways the tool can be invoked)
	Tags        []string // meta data about tool for users
}

// The actual agent information (ties together modal settings, tools, and user info for agents)
type Agent struct {
	AgentId        string        // unique agent id
	UserInfo       *User         // avatar
	Settings       ModelSettings // settings and model llm
	ToolsAvailable []SystemTool  // tools the agent can use
}

// Algorithms for gaining consensus in a multi-agent conversation
// regarding a specific task. These are encoded as scripts so they
// can be easily configured and reconfigured at runtime
type Konsensus struct {
	Id             string    // unique id
	Name           string    // Name of algo
	InstructionSet *Resource // The resource identifier for the script
}

// Instead of simply prompting, we will "work with" the user to attempt
// to clarify and extract as much detail relevant to the task as possible
type StatementOfIntent struct {
	Id                    string
	InitialStatement      string   // What the user said they wanted
	RefinedStatement      string   // Carified document from user/ai interaction
	RefiningSteps         []string // Refinement overview, statements from AI ABOUT the refinement process
	ExplicitExpectations  []string // A list of explicit expectations from the user
	PointsOfConcernListed []string // a list of primary points of concern regarding the task
}

type TopLevelCategory struct {
	Id                     string    // unique
	Name                   string    // unqiue
	Description            string    // optional
	RefinementInstructions *Resource // script to "refine" the selected model to be better suited for task
}

// A user project
type Project struct {
	Id            string
	Name          string
	Soi           StatementOfIntent
	Agents        []*Agent
	NonAgentUsers []*User
	Resources     []*Resource
	Sessions      []*Session
}
