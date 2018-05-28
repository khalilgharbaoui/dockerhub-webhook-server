# Dockerhub Webhook Server

Add this to your project on the server and use `go build dockerhub-webhook-server.go`.

On your server do: `export WEBHOOK_AUTH_TOKEN="X12349999"`

Then run it with `nohup ./dockerhub-webhook-server`

(you need to install nohup)

It will spin up a server listening on port `8008`

It give you a webhook url like: http://domain-or-ip:8008/X12349999

Use it for example in Dockerhub, so each time you push your image to the registry it will execute the script in the folder `.webhook/setup`.

Edit the Script to your liking!.

In the moment it stands like this:

```
  #!/bin/sh
  echo "Stop..."
  docker-compose stop web
  docker-compose stop worker
  echo "Pull..."
  docker pull kaygeee/inquiries-maker:latest
  echo "Remove..."
  docker rm inquiries-maker_web_1
  docker rm inquiries-maker_worker_1
  sleep 1
  echo "Build..."
  docker-compose up -d
  echo "Done!"

```

- It stops my web and worker containers.
- Pulls the new image from the registry.
- Removes the stopped containers.
- builds them up with the new image.
- without disabling the database or any other containers.

You can see the logs of the setup script by opening a new terminal window on your server and using: `tail -f nohup.out`

I strongly recommend you keep this file structure and add it to the root of your project where your docker-compose.yml file lives.

You can test it locally without adding it anywhere:
`curl http://domain-or-ip:8008/X12349999`
