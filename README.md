# Commands Runner
This project allows you to orchestrate commands by creating a descriptor file which is an array of states. Each state have a command to run, a status, a log locatio, a status which can have values (READY, SKIP, RUNNING, SUCCEEDED and FAILED) and other parameters. The command runner will run the descriptor file and mark the state as SUCCEEDED or FAILED depending on the exit code of the command. You can also extend this project with your own specific requirements.<br>

## Getting Started
You can run the `commands runner` by installing it as a server or by calling it programmatically. You can also extend the server if needed.

### Pre-requisites
This project uses `dep` to manage dependencies.<br>
1. Install [https://github.com/golang/dep/blob/master/docs/installation.md] 
2. Create your project
3. Run `dep init`
4. In the newly created `Gopkg.toml` file add a `constraint` section. (once the code will be migrated to github.com a simple `dep ensure -add will replace this insertion)
```
[[constraint]]
  name = "github.ibm.com/IBMPrivateCloud/cfp-commands-runner"
  source = "git@github.ibm.com:IBMPrivateCloud/cfp-commands-runner.git"
```
1. run `dep ensure -v`, this will download all dependencies.

### Create a commands-runner server
1. Create server: There a server example at [examples/server](./examples/server). In that example the server is enriched with a `helloWorld` API.
2. Build server: Once you created the server, you can build it with for example: `go build -o server  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/server`.
3. Create certificates (optional): You can secure the communication between the client and the server using SSL.
  1. `openssl req -x509 -newkey rsa:4096 -keyout <your_data_directory>/cr-key.pem -out  <your_data_directory>/cr-cert.crt  -days 365 -subj "/C=YourContry/ST=YourState/L=YourLocation/O=YourOrg/OU=YourOrgUnit/CN=localhost" -nodes`
  2. Install the certificate on the machine which will run the server.
  For example on Ubuntu or MaxOS:
    1. `cp <your_data_directory>/cr-cert.crt /usr/local/share/ca-certificates/`
    2. `update-ca-certificates`
4. Launch the server: run the command `./server listen -c <your_data_dir> -s <your_states_file` (see below for details on states_file).<br>

A state file example is provided in the [examples/data](./examples/data).

### Create a commands-runner client
1. Create the client: There a client example at [examples/client](./examples/client). In that example the client is enriched with a command `hello` which call the `helloWorld` API on the server side.
2. Build the client:  Once you created the client, you can build it with for example: `go build -o client  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/client`.
3. Create token: The server uses a token for authentication, run the command: `./client token create > <your_data_directory>/cr-token`, this will create a file `cr-token` in `<your_data_directory>`.
4. Setup the client: `./client --url <server_url> --token <token> --cacert <cert_path> api save` and finally use it.
5. launch `./client` for more information on all available commands.

### Use commands-runner in a program.
There is code examples at [examples/code](./examples/code)

## Sever Overview
A states file is a yaml file which describes the states to execute. You can find an [examples/data/states.yml](./examples/code/states.yml). 
The states are executed sequentially but you can alter the sequence by adding for each state what should be the next state, 
in that case a preprocessing will be done to re-order the states in a topological order. At the first run all states are marked either READY or SKIP, after the first execution the states can be marked as "SUCCEEDED" (if the execution SUCCEEDED), "FAILED" (if the execution failed), 
"SKIP" (if the state must be skipped) or "READY" (if the state is not executed and probably because a previous state failed). When the command runner starts the execution of the states file and it runs each state marked READY or FAILED, if the state execution failed then the state will be marked FAILED and the execution stops otherwise the state is marked "SUCCEEDED" and the execution continues. If a state is marked as SUCCEEDED or "SKIP" it will be not executed and the command runner will start the execution a the next READY or FAILED state.<br>

### States file format:

```
states: An array of state
- name: name of the state
  phase: If the value is "AtEachRun" then this state will run each time whatever the status except if "SKIP"
  label: label of the state
  log_path: log path, the stdin/out of the script command will be forwared in that file.
  status: status "READY", "SUCCEEDED", "RUNNING", "FAILED", "SKIP"
  start_time: The last start time the state ran
  end_time: The last end time the state ran
  reason: The reason of failue
  script: The command to execute, it can be an absolute path or a path relative to location of the state files.
  script_timeout: The timoute for executing that state
  protected: If true then the state can not be removed using the client CLI
  deleted: If true then the state will be deleted at the next merge between the old states file and the new state file.
  states_to_rerun: An array of states (name) to rerun once this state is executed. The states to rerun must be placed after the current state in the topological order.
  next_states: An array of the next states to run after this one.
  previous_states: This is calculated array and so every information set here will be overwritten by the command runner.
- name:
  ...
```
### Send config file to the server
You can send a config file to the server before starting the processing of the state engine and this using the client command:
```./client config  save -c <config_file_path>```

<a name="configFileFormat"></a>
The config file is a yaml file in the form of:

```
config:
  property1: hello world
  ...
```

The root attribute `config` is configurable using `configManager.SetConfigYamlRootKey("myconfig")` along with the config file name `configManager.SetConfigFileName("myconfig.yml")` (see: [examples/server/server.go](./examples/server/server.go))

### Launch the commands-runner
Once the server is up and running with your states file, you can launch the commands-runner using the command:
```./client engine start```

You can check the progress using command: 
```./client logs -f```

The command runner works as follow:<br>

1. Read the state files
2. Stops if one of the state has a status RUNNING
3. Starts the execution at the first state with status READY or FAILED, if the state is an extension then the commands runner will start the first state (READY or FAILED) of the extension process.
4. If the state failed, the commands runner stops.
5. If the state succeeded, the commands runner search for the next READY or FAILED status to execute.
6. If no more state to process then the commands runner stops.


### Extensions

You can insert extension in a states file using the client CLI or API. An extension is an artificat which contains a manifest describing (sub-)states to execute and where to insert this execution in the states file. Inserting an extension will create a new state in the states file and this new state will call the extension process and run it using the commands-runner.<br>

The extension can be either "embedded" or "customer", "embeded" means that the artifact is already provided in the environement (ie: part of the product distribution) and should not be register and can not be deleted. A "embeded" extension is defined in a extensions yaml file as in [examples/data/test-extensions.yml](./examples/data/test-extensions.yml). A "custom" extension must be registered using the client CLI and for that it must be embobined in a zip file.<br>

Once the extension is in the environmeent (register), the exentions must be registered and then inserted in the states file and so when the commands-runner will be launched, the extension will get executed along the states file.

Extensions can be inserted in an extension states file once the parent extension is registered.

#### Extension directory structure

An simple example is provided here [examples/extenions/simple-extension-with-version](./examples/extenions/simple-extension-with-version) or [examples/extenions/simple-extension-without-version](./examples/extenions/simple-extension-without-version). The version is an intermediate directory under the extension itself, it allows to see in the reposity which version gets deployed. The commands-runner doesn't support multiple versions in the repository for the time being. The only constraint is that the extension must have a `extension-manifest.yml` in the extenision root directory.

#### Extension manifest format

The extension-manifest is a yaml file. Only one attribute is mandatory `states` which is an array of state. The state structure can be found at [api/commandsRunner/stateManager/stateManager.go](./api/commandsRunner/stateManager/stateManager.go).<br>
Manifest examples can be found in the [examples/extensions](./examples/extensions)

#### How to setup the server to manage extension

In [examples/server](./examples/server) example you can see that the extensionManager is initialized with files and directory values.

#### Embeded extension

Embeded extension are provided by your distribution and so the "end-user" doesn't need to register them into the environment. They are automatically registered based on the content embeddedExtensionDescriptor provided in the extensionManager.Init method.
You can find an example of this file at [examples/data/test-extensions.yml](./examples/data/test-extensions.yml).

#### Register custom extensions

If the end-user wants to register its own extension, he must create a zip file containing the extension structure.<br> 
For example by executing: 
```zip simple-custom-extension.zip simple-custom-extension/*```

and then use the client CLI: 
```./client extension -e <extension_name> register -p <archive_path>```

#### Send extension's configuration in the environment


Like for the main state file, You can send a file (configuration) to the environemnt using:

```./client extension -e <extension_name> save -c <config_file>```

The config will be saved in the extension directory.

Format see: [config file format](#configFileFormat)

#### Insert extensions

As the previous state and next state are defined in the `call_state` attribute of the `extension_manifest.yml`, to insert the extension in the states file, you have to execute.

```./client states insert -i <extension_name>```


