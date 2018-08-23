# --------------------------------------
# Build Stage
# --------------------------------------
FROM golang:1.10 as build_stage
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN go build -o main . 

# --------------------------------------
# Production Container
# --------------------------------------
FROM scratch
COPY --from=build_stage /app/billmonger /
CMD [/billmonger]
