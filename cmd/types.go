package cmd

type cliConfig struct {
	ProfileFlag       *string
	CredentialsFile   *bool
	RegionFlag        *string
	Retain            *int8
	Verbose           *bool
	DryRun            *bool
	LambdaListFile    *string
	MoreLambdaDetails *bool
	SizeIEC           *bool
}
