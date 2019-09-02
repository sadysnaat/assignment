# build stage
FROM golang
WORKDIR /work
ADD . .
RUN go build -o /bin/assignment .
WORKDIR /
RUN rm -r /work
ENTRYPOINT /bin/assignment