MyLAPS streamer
------------

Instructions:

```bash
go build -o main
./main
```

Deploy:

```bash
gcloud components install app-engine-go
gcloud app deploy app.yaml --project PROJECT_ID
```
