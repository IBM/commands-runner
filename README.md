# Commands Runner
This project allows you to orchestrate commands by creating a descriptor file which is an array of states. Each state have a command to run, a status, a log location, which other states have to be reprocess in case of error and other parameters. The command runner will run the descriptor file and mark the state as SUCCEEDED or FAILED depending on the exit code of the command. You can also extend this project with your own specific requirements.<br>

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

### Use commands-runner in a program.
There is code examples at [examples/code](./examples/code)

## Sever Overview
A states file is a yaml file which describes the states to execute. You can find an [examples/data/states.yml](./examples/code/states.yml). 
The states are executed sequentially but you can alter the sequence by adding for each state what should be the next state, 
in that case a preprocessing will be done to re-order the states in a topological order. At the first run all states are marked either READY or SKIP, after the first execution the states can be marked as "SUCCEEDED" (if the execution SUCCEEDED), "FAILED" (if the execution failed), 
"SKIP" (if the state must be skipped) or "READY" (if the state is not executed and probably because a previous state failed). When the command runner starts the execution of the states file and it runs each state marked READY or FAILED, if the state execution failed then the state will be marked FAILED and the execution stops otherwise the state is marked "SUCCEEDED" and the execution continues. If a state is marked as SUCCEEDED or "SKIP" it will be not executed and the command runner will start the execution a the next READY or FAILED state.<br>

### State file format:

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

### Extensions
You can insert extension in a states file using the client CLI or API. An extension is an artificat which contains a manifest describing (sub-)states to execute and where to insert this execution in the states file. Inserting an extension will create a new state in the states file and this new state will call the extension process and run it using the commands-runner.<br>

The extension can be either "embedded" or "customer", "embeded" means that the artifact is already provided in the environement (ie: part of the product distribution) and should not be loaded and can not be deleted. A "embeded" extension is defined in a extensions yaml file as in [examples/data/test-extensions.yml](./examples/code/test-extensions.yml). A "customer" must be loaded using the client CLI and for that it must be embobined in a zip file.<br>

Once the extension is in the environmeent (loaded), the exentions must be registered and then inserted in the states file and so when the commands-runner will be launched, the extension will get executed along the states file.

Extensions can be inserted in an extension states file once the parent extension is registered.

#### Extension directory structure

An simple example is provided here [examples/extenions/simple-extension](./examples/extenions/simple-extension). The only constraint is that the extension must have a `extension-manifest.yml` in the extenision root directory.

#### Extension manifest format

The extension-manifest is a yaml file. Only one attribute is mandatory `states` which is an array of state. The state structure can be found at [api/commandsRunner/stateManager/stateManager.go](./api/commandsRunner/stateManager/stateManager.go).<br>

#### How to setup the server to manage extension

In [examples/server](./examples/server) example you can see that the extensionManager is initialized with files and directory values.

#### Load custom extension

#### Register extensions

#### Insert extensions




