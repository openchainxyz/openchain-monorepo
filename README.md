# OpenChain Monorepo

This monorepo contains all the code which powers [OpenChain](https://openchain.xyz)

## Requirements
[Bazel](https://bazel.build/install)
[Go](https://go.dev/doc/install)
## Getting Started
## Easy Method
Install the dependancies by running
```
bash build.sh
```
Then you run this to start up the service
```
bash run.sh
```
## Harder Method
Install the dependancies by running
```
bazel run cmd/signature-database-srv/BUILD.bazel
```
Then you run this to start up the backend
```
bazel run cmd/signature-database-srv/main.go
```

You can run the frontend like this
```
pnpm install
pnpm run dev
```
## After installing and running
Go to localhost:3000 in your web browser. Keep in mind that if port 3000 is taken, it will default to port 3001.