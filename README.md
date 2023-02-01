# Setup
Prerequisites: docker, docker-compose, go
1. git clone github.com/ws396/autobinance
2. cp .env.example .env
3. docker compose up
4. go run .

This will start the SSH server to which you can connect by typing `ssh localhost -p 23234`, optionally changing the IP and/or port.

After setting up you need to specify trading strategies and symbols (ex. example LTCBTC) and then launch the trading session.

Please keep in mind that this software should only be used for educational purposes.
