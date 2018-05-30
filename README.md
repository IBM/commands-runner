# Commands Runner
This project allows you to orchestrate commands by creating a descriptor file which is an array of steps. Each step have a command to run, a status, a log location, which other steps have to be reprocess in case of error and other parameters. An engine will run the descriptor file and mark the step as SUCCEEDED or FAILED depending on the exit code of the command. You can also extend this project with your own specific requirements.<br>

## Getting Started
You can run the `commands runner` by installing it as a server or by calling it programmatically. You can also extend the server if needed.

### Pre-requisites
This project uses `glide` to manage dependencies.<br>
1. Install [https://github.com/Masterminds/glide] 
2. Create your project
3. Run `glide create`
4. In the newly created `glide.yml` file add in the `import` section.
```
- package: github.ibm.com/IBMPrivateCloud/cfp-commands-runner
  repo: git@github.ibm.com:IBMPrivateCloud/cfp-commands-runner.git
  subpackages:
  - commandsRunner
  - commandsRunnerClient
```
5. run `glide update`, this will download all dependencies.

### Create a commands-runner server
There a server examples at [examples/server](./examples/server). In that example the server is enriched with a `helloWorld` API.<br>
Once you created the server, you can build it with for example: `go build -o server  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/server` and then run it with `./server listen -c <your_data_dir> -s <your_states_file`.<br>
A state file example is provided in the [examples/data](./examples/data).

### Create a commands-runner client
There a server examples at [examples/client](./examples/client). In that example the client is enriched with a command `hello` which call the `helloWorld` API on the server side.<br>
Once you created the client, you can build it with for example: `go build -o client  github.ibm.com/IBMPrivateCloud/cfp-commands-runner/examples/client` and then setup it with `./client api....` and finally use it.

### Use commands-runner in a program.
There is code examples at [examples/code](./examples/code)

### Installing

## 