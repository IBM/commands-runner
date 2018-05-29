# Commands Runner
This project allows you to orchestrate commands by creating a descriptor file which is an array of steps. Each step have a command to run, a status, a log location, which other steps have to be reprocess in case of error and other parameters. An engine will run the descriptor file and mark the step as SUCCEEDED or FAILED depending on the exit code of the command. You can also extend this project with your own specific requirements.

## Getting Started
This project uses `glide` to manage dependencies.<br>
1. Clone this project.
2. Install [https://github.com/Masterminds/glide] 
    1. in Mac or Linux run `make glide-install`
    2. On other system follow the documentation
3. run `make pre-req` in order to download all depdencies.

### Prerequistes

### Installing

## 