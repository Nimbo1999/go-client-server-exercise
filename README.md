<div>
    <h1 align="center">Client-Server API exercise</h1>
    <p align="center">Exercise of database module from <a href="https://goexpert.fullcycle.com.br/pos-goexpert/" target="_blank">Go Expert</a> course from <a href="https://fullcycle.com.br/" target="_blank">FullCycle</a><p>
</div>
<hr>

### Prerequisite to run the apps.

You must have [Docker](https://docs.docker.com/engine/install/) and [GOlang](https://go.dev/dl/) installed into your machine.

### How to run the project into your machine.

1. Clone the repository by running the command:
    ```(bash)
    git clone https://github.com/Nimbo1999/go-client-server-exercise.git
    ```

2. Access the project directory:
    ```(bash)
    cd go-client-server-exercise
    ```

3. Run docker compose to generate the test.db SQLite file:
    ```(bash)
    docker-compose up -d
    ```

4. Run the server by executing the following command in your terminal:
    ```(bash)
    cd ./server && go mod tidy && go run .
    ```

5. Run the client app by executing the command:
    ```(bash)
    cd ./client && go mod tidy && go run .
    ```

