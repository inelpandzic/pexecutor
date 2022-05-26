# Pexecutor - Simple Toy Task Executor


## Run

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
    {"name": "Task 10","duration":4000},
    {"name": "Task 11","duration":5100}]'

```

```
curl localhost:8080/tasks/running | jq .
```

```
curl localhost:8080/tasks/pending | jq .
```
