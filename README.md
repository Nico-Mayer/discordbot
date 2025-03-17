# Start Local Container

1. docker build -t discordbot .
2. docker run -d --name discordbot --env-file .env.prod discordbot
