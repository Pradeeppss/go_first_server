FROM golang:1.19

WORKDIR /app

COPY . .

RUN ["go","build","wiki.go"]

CMD [ "./wiki" ]

EXPOSE 8080