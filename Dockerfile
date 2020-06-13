FROM golang:1.14

# Update and install curl
RUN apt-get update

# Creating work directory
RUN mkdir /code

# Adding project to work directory
ADD . /code

# Choosing work directory
WORKDIR /code

# build project
RUN go build -o pepe_steam_bot .

CMD ["./pepe_steam_bot"]