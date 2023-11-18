aws-regions.txt:
	aws ec2 describe-regions --region us-east-1 --all-regions --query "Regions[].RegionName" --output text > cmd/aws-regions.txt

license:
	@echo "Applying license headers..."
	 copywrite headers

opensource:
	@echo "Checking for open source licenses"
	~/go/bin/go-licenses report github.com/karl-cardenas-coding/go-lambda-cleanup --template=documentation/open-source.tpl > documentation/open-source.md 


lint: ## Start Go Linter
	@echo "Running Go Linter"
	golangci-lint run  ./...


build: ## Build the binary file
	@echo "Building lambda function and adding version number 1.0.0"
	 go build -ldflags="-X 'github.com/karl-cardenas-coding/go-lambda-cleanup/cmd.VersionString=1.0.0'" -o=glc -v 

tests: ## Run tests
	@echo "Running tests"
	go test -race -shuffle on ./...


tests-coverage: ## Start Go Test with code coverage
	@echo "Running Go Tests with code coverage"
	go test -race -shuffle on -cover -coverprofile=coverage.out  ./...

view-coverage: ## View the code coverage
	@echo "Viewing the code coverage"
	go tool cover -html=coverage.out