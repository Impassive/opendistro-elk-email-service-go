FROM golang
ADD . ./src/goAPI
RUN ls -la ./src/goAPI/ && go install goAPI && cp -pr ./src/goAPI/tpl ./bin
ENTRYPOINT ["/go/bin/goAPI"]
EXPOSE 8080