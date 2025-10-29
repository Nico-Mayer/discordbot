# Start Local Container

1. docker build -t discordbot .


2. Run contianer
- prev
  docker run -d --name discordbot--prev --env-file .env.prev discordbot
- prod
  docker run -d --name discordbot--prod --env-file .env.prod discordbot
