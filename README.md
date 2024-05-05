# gitlab-issuereporter
gitlab-issuereporter is a tool used to check for blocker bugs in staging/production

# Usage
gitlab-issuereporter will return a excel sheet as output with issues containing assignee , date with description of issue /ticket number

To filter by a release label and "READY-FOR-TEST":
```shell
go run main.go -release "Release::Mar-26-2024" -ready-for-test
```

To filter by a release label and a specific blocker label:
```shell
go run main.go -release "Release::Mar-26-2024" -blocker "blocker/staging-upgrade"
```shell

To filter by a release label, "READY-FOR-TEST", and a specific blocker label:
```shell
go run main.go -release "Release::Mar-26-2024" -ready-for-test -blocker "blocker/production-upgrade"
```shell
If we didn't Specify "ready-for-test" in CLI , it automatically picks as NOT Ready For Test.