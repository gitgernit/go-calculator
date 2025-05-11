# go-calculator
Project for the Yandex.Lyceum Go programming course.  

## Prerequisites
* go
* Docker (optionally)

## Running
2 options are present, stick to the one that fits you more:

1. Run directly through go: `go run cmd/[orchestrator,agent]/**/main.go`
2. Use docker with compose plugin: `docker compose --env-file=./configs/.env up --build`

Docker compose will expect you to have some environment variables, 
hence, you'll need to create an .env file or export them manually. 
Feel free to create the .env file based on ./configs/.env.template

## Workflow
go-calculator uses 2 services -- "orchestrator" and "agent"

The orchestrator service manages expressions through the following user interface:
* Create an expression
* Get expression
* List all expressions

Orchestrator also provides an interface for the agent, whom will be covered later:
* Get next task
* Submit task result

### Orchestrator
Orchestrator is an HTTP service that lets you input mathematical expressions,
leave the evaluation on behalf of the orchestrator, and so, get expressions
results. Orchestrator itself doesn't solve the expressions - instead, 
the orchestrator converts given expressions from infix form to 
reverse polish notation form, and depends on external services to
solve the expressions step-by-step. In this case, "agent" does the solving part.

### Agent
Agent is a daemon which uses polling to fetch "tasks" from the orchestrator.
A task is essentially an expression in RPN. Solving a task means solving
only one step, not the full RPN. Agent utilizes parallelism to solve
multiple tasks concurrently. RPN is fully dependant of stack order,
thus, one task can be assigned to only one poller at a time.
If an expression has 2 operators, each taking 2 seconds to evaluate,
the expression will still take 4 seconds to be fully solved because of RPN limitations.

## Environment variables
```
ORCHESTRATOR_HOST - self-explanatory
ORCHESTARTOR_PORT - self-explanatory
ORCHESTARTOR_GRPC_PORT - self-explanatory

TIME_ADDITION_MS - "+" operator time complexity
TIME_SUBTRACTION_MS - "-" operator time complexity
TIME_MULTIPLICATIONS_MS - "*" operator time complexity
TIME_DIVISIONS_MS - "/" operator time complexity

COMPUTING_POWER - amount of concurrent agent pollers
POLLING_INTERVAL_MS - interval for pollers to fetch tasks between
```
