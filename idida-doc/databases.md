# Guide to Using the Database and Related Managers

## Database (database.go)

The `Database` struct is the core component for interacting with the database. It provides CRUD operations for various entities:

- Resource
- TokenData
- Token
- Passport
- User
- Message
- Session
- Conversation
- ModelSettings
- SystemTool
- Agent
- Konsensus
- StatementOfIntent
- TopLevelCategory
- Project

Each entity has four basic operations:
- Create: `CreateX`
- Read: `GetXByID`
- Update: `UpdateX`
- Delete: `DeleteX`

The `NewDatabase` function is used to create a new database instance or load an existing one.

## AgentManager (agentdata.go)

The `AgentManager` struct provides operations for managing agents:

- `CreateAgent`: Creates a new agent
- `GetAgent`: Retrieves an agent by ID
- `UpdateAgent`: Updates an existing agent
- `DeleteAgent`: Deletes an agent
- `AddToolToAgent`: Adds a SystemTool to an agent

## User (user.go)

The `NewUser` function is used to create a new User with validation for username, password, and email.

## ConversationManager (convo.go)

The `ConversationManager` struct handles operations related to conversations:

- `CreateConversation`: Creates a new conversation
- `GetConversation`: Retrieves a conversation by ID
- `UpdateConversation`: Updates an existing conversation
- `DeleteConversation`: Deletes a conversation
- `AddParticipantToConversation`: Adds a user to a conversation
- `AddMessageToConversation`: Adds a message to a conversation

## Usage Pattern

1. Initialize the database using `NewDatabase`.
2. Create manager instances (e.g., `NewAgentManager`, `NewConversationManager`) using the database instance.
3. Use the managers to perform operations on their respective entities.
4. For entities without specific managers, use the CRUD operations provided by the `Database` struct directly.

This structure allows for organized and type-safe interactions with the database, separating concerns between different entity types and providing a clear interface for each operation.
