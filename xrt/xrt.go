/*
Runtime will be the structure that actually plays as like a service or a server and rebels and interactions will be connected to it locally it will also be the thing that vines to the ollama instances that we will use to control the agents and it will be pretty much how we store all memory data everything will be the core nexus of the xla

We need some sort of directory that points to users each user will have their own databases there are agents those databases can be vector and also will have vector database stored for the user fully attend a flat file structure for the project as we start
I really don't play with databases a whole bunch but I may just use SQL light and plug those user and user places just so we don't have to write a bunch of code to support file system operations and stuff
*/
package xrt

const (
	SettingPrimaryServiceDbName = "xla.service.db"
	SettingPrimaryUserDbName    = "xla.db"   // That user's Agent config/ settings/ etc
	SettingDocumentStore        = "xla-docs" // directory that will contain a sub-directory for all local users' project documents and meta data/ project-specific vector databases
)

type Runtime struct {
	XrtRootDir string // this will be the route directory that will play as the install directory for XLA

}

/*
TODO:		create a function that will build a new runtime given a route directory it may return an error if the route directory does not exist or if a bullion given for create is not given.
If the bullion create is given the directory must not exist we will create the directory with all default settings as described below.
If the bullion flag is set to false as and don't create, the path does exist as a directory then we need to load the contents of the director directory to start the runtime service.
That means loading and running any migration operations required on xla.service.DB which will be a GORM database ORM SQL light three

 the runtime configuration must include in a string that represents an O operating system environment variable that will be read from potentially any operating system windows back clinics.
It is imperative that we check the environment for this variable as it will be the route password for the admin account for the user interface that handles the web interaction to the runtime
the os environment variable name default SHALL be : XLA_ADMIN_WEB_TOKEN if this token is not sent them it should exit with an error

When a new runtime is being created we also need to create TLS certificates so we can serve HTTPS, we will generate these certificates into the directory and they will be self signed
The runtime configuration shall specify TLS cert and key, which will be path to those files this way if a user wants to use their own Certs they can just update the paths and restart the runtime

TODO:

	 the runtime shall use gin https only. TLS certificates cannot be found or loaded the program must not run.
	 There needs to be a date for JW tokens that we will make and create similar to the NC labs badger type
	 The runtime web administration interface needs to be created where new access tokens and users can be created in removed
	 We also need to have default users that are created or a set up script that quest the that users be set up

	 Now we don't need to figure out how runtimes we talk to each other across systems yet because these are still intended to be local installs
	 However we do need to ensure that we can specify the ollama host and binding to the server as the ollama server may be remote
	 Also keep in mind that we need to interface things off such that we're not bound to ollama, that will be the intended target so we may tight couple


TODO:
	Need to write the xlist evaluator which will send a call back, or rather, it will look for the appropriate action to take given the list context it is in

	I would like to also create a tag added to the collector and everything that prefix with an at symbol or something so we can specify a file
	Somewhere on the users file system that can be used to be vectorized brought in through whatever means necessary
	My idea is that there would be some sort of hold on some sort of stream source
	That comes in just a memory memory hog just full Java script(metaphor)
*/
