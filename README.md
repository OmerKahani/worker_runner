# worker_runner
run cmd worker in k8s with http healthcheck,
Ther server will serve healthcheck until the worker finish and exit.
SIGTERM will be delegated to the worker

example: rake jobs:work

Usages:
```
worker_runner rake jobs:work
```

