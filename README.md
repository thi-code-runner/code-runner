# CODE-RUNNER

## DESCRIPTION
code-runner is a sandbox capable of running arbitrary 
source code by using Docker containers.
It allows for interaction between client and server by using the
WebSocket protocol. Code-runner is configured with a configuration file
which describes how code has to be compiled and run.

## CONFIGURATION

example configuration file:
````
{
  "hostCleanupIntervalS": 360,
  "cacheCleanupIntervalS": 180,
  "containerConfig": [
    {
      "id": "java-20",
      "image": "eclipse-temurin:20-jdk",
      "compilationCmd": "javac *.java",
      "executionCmd": "java {{getSubstringUntil .FileName \".\"}}",
      "reserveContainerAmount": 0,
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
      "reserveContainerAmount": 0,
      "readOnly": false
    }
  ]
}

````
- id: correlates request with config block
- image: Docker image
- compilationCmd: compilation command
- executionCmd: execution command (template method getSubstringUntil can be used)
- add: what resources to add to the container
- reportExtractor: which extractor to use (uses stdout if no reportPath is configured)
- reportPath: which path the test report is written to (inside the container)
- Memory: container memory limit
- CPU: container CPU limit
- reserveContainerAmount: start containers beforehand to save on startup time
- readOnly: wether file access should be possible
- diskSize: size of disk if readOnly is set to false

## API

example API payload:
```
{
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
types of WebSocket messages:
- execute/run
- execute/input
- execute/test

structure of the payload can be examined in the model folder