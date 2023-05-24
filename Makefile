aws-regions.txt:
	aws ec2 describe-regions --region us-east-1 --all-regions --query "Regions[].RegionName" --output text > cmd/aws-regions.txt

license:
	@echo "Applying license headers..."
	 copywrite headers


opensource:
	@echo "Checking for open source licenses"
	~/go/bin/go-licenses report github.com/karl-cardenas-coding/go-lambda-cleanup --template=documentation/open-source.tpl > documentation/open-source.md 