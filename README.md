# Pexecutor - Simple Task Executor

Pexecutor is a simple web task executor. It allows you to submit tasks that will be executed by the first available worker. Task that are already running or are waiting to be executed will be rejected as duplicates.

## Run

Requirements: `go 1.17` available

Run the service:
```
go run main.go -pool-size 5 - queue-size 1000 -port 8080
```
This will startup a service with the executor of 5 workers and task queue capacity of 1000.


## Usage

Submit couple of tasks:

```
curl -X POST localhost:8080/tasks \
   -H "Content-Type: application/json" \
   -d '[{"name": "Task 1","duration":3000}, 
    {"name": "Task 2","duration":2000},
    {"name": "Task 3","duration":5000},
    {"name": "Task 4","duration":7000},
    {"name": "Task 5","duration":7500},
    {"name": "Task 6","duration":2200},
    {"name": "Task 7","duration":8000},
    {"name": "Task 8","duration":3330},
    {"name": "Task 9","duration":2200},
    {"name": "Task 10","duration":5100}]'

```

And the response can be something like:
```
{
  "RequestedTask": 10,
  "SubmittedTasks": 7,
  "DuplicateTasks": 3 
}
```
This means that 3 tasks are rejected(duplicate) because they are executing or waiting to be executed.


Get the running tasks:
```
curl localhost:8080/tasks/running | jq .
[
  {
    "name": "Task 6",
    "duration": 2200
  },
  {
    "name": "Task 10",
    "duration": 4000
  },
  {
    "name": "Task 7",
    "duration": 8000
  },
]

```


Or get the pending tasks like:
```
curl localhost:8080/tasks/pending | jq .
[
  {
    "name": "Task 8",
    "duration": 3330
  },
  {
    "name": "Task 9",
    "duration": 2200
  }
]
```

## Things to Improve

This is the list of known things that can be improved:
- Better and more tests.
- If the task queue is exhausted, submitting a task will be blocked and with the all workers busy with long running tasks, the system can halt. The end user should not wait necessarily for this operation.
- Workers should be gracefully shutdown.
- Using a concurrent map like [orcaman/concurrent-map](https://github.com/orcaman/concurrent-map) for the backing map will be more efficient because it is sharded and only the shard to which the key belongs iis locked, rather than the whole map like it is done now.
- It would be nicer if it is packaged in a Docker container

