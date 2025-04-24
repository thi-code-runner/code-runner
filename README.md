# CODE-RUNNER

## Description

code-runner is a sandbox capable of running arbitrary source code by using Docker containers.
It allows for interaction between client and server by using the WebSocket protocol.
Code-runner is configured with a configuration file which describes how code has to be compiled and run.

## Configuration

Example configuration file:

```json
{
  "hostCleanupIntervalS": 360,
  "cacheCleanupIntervalS": 180,
  "containerConfig": [
    {
      "id": "java-20",
      "image": "eclipse-temurin:20-jdk",
      "compilationCmd": "javac *.java",
      "executionCmd": "java {{getSubstringUntil .FileName \".\"}}",
      "Memory": 100,
      "CPU": 0.5,
      "readOnly": true
    },
    {
      "id": "junit-5",
      "image": "eclipse-temurin:20-jdk",
      "compilationCmd": "javac -d target -cp target:resources/junit-platform-console-standalone-1.9.3.jar *.java",
      "executionCmd": "java -jar resources/junit-platform-console-standalone-1.9.3.jar --class-path target --select-class {{getSubstringUntil .FileName \".\"}} --reports-dir=./reports --details=tree",
      "add": ["./resources/junit-platform-console-standalone-1.9.3.jar"],
      "reportExtractor": "junit-5-out",
      "Memory": 100,
      "CPU": 2,
      "readOnly": false
    }
  ]
}
```

- id: correlates request with config block
- image: Docker image
- compilationCmd: compilation command
- executionCmd: execution command (template method getSubstringUntil can be used)
- add: what resources to add to the container
- reportExtractor: which extractor to use (uses stdout if no reportPath is configured)
- reportPath: which path the test report is written to (inside the container)
- Memory: container memory limit
- CPU: container CPU limit
- readOnly: wether file access should be possible
- diskSize: size of disk if readOnly is set to false

## Build

```bash
make build
```

... or ...

```bash
docker build -t code-runner .
```

## Start

```bash
CODE_RUNNER_CONFIG=`pwd`/config.json ./bin/code-runner
```

... or ...

```bash
docker run --rm -p 8080:8080 -v `pwd`/config.json:/etc/code-runner/config.json -v /var/run/docker.sock:/var/run/docker.sock code-runner
```

... or ...

```bash
docker compose up -d --build
```

## API

Calls can be tested by using any websocket based client. One would be `wscat`, can be installed with `npm install -g wscat`. Connect to server by using:

```bash
wscat -c ws://localhost:8080/run
```

### Run: execute/run

Example API payload:

```json
{
  "id": "5a37fbe9-6fed-47f9-a42f-add804e4e82c",
  "type": "execute/run",
  "data": {
    "cmd": "java-20",
    "mainfilename": "Main.java",
    "sourcefiles": [
      {
        "filename": "Main.java",
        "content": "class Main{public static void main(String[] args) {System.out.println(\"Hello World!\");}}"
      }
    ]
  }
}
```

### Input: execute/input

Add input, if example waits for input.

```json
{
  "id": "3b53877b-792c-45a8-ac25-cb8605a57a2d",
  "type": "execute/run",
  "data": {
    "cmd": "java-20",
    "mainfilename": "Main.java",
    "sourcefiles": [
      {
        "filename": "Main.java",
        "content": "class Main{public static void main(String[] args) throws Exception { System.in.read();System.out.println(\"Hello World!\"); }}"
      }
    ]
  }
}
```

```json
{
  "type": "execute/input",
  "stdin": "\n"
}
```

### Test: execute/test

Run test with test framework, e.g., JUnit or Output compare

```json
{
  "id": "d385e01a-34af-454c-9afa-4ef7cc74c3f9",
  "type": "execute/test",
  "data": {
    "cmd": "java",
    "mainfilename": "Main.java",
    "tests": [
      {
        "type": "output",
        "param": {
          "expected": "Hello World!"
        }
      }
    ],
    "sourcefiles": [
      {
        "filename": "Main.java",
        "content": "class Main{public static void main(String[] args) {System.out.print(\"Hello World!\");}}"
      }
    ]
  }
}
```

_Hint: Structure of the payload can be examined in the model folder_
