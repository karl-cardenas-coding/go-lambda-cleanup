#!/bin/bash

# Replace these variables with your own values
BASE_FUNCTION_NAME="function1"
AWS_REGION="us-east-1"

export AWS_PAGER=""

# Loop to update, publish new versions, and create aliases
for ((i=1; i<=2000; i++)); do
    # Create a new temporary directory
    TEMP_DIR=$(mktemp -d)
    ZIP_FILE_PATH="${TEMP_DIR}/handler.zip"

    # Add a dummy comment to the ZIP file
    echo "# Dummy comment for update" > "${TEMP_DIR}/handler.js"

    # Zip the contents of the temporary directory
    zip -j $ZIP_FILE_PATH "${TEMP_DIR}"/*

    # Update Lambda function code
    FUNCTION_NAME="${BASE_FUNCTION_NAME}"
    aws lambda update-function-code --function-name $FUNCTION_NAME --zip-file fileb://$ZIP_FILE_PATH --region $AWS_REGION --profile automation

    # Wait for the update to complete
    aws lambda wait function-updated --function-name $FUNCTION_NAME --region $AWS_REGION --profile automation

    # Publish a new version
    NEW_VERSION=$(aws lambda publish-version --function-name $FUNCTION_NAME --region $AWS_REGION --profile automation --query 'Version' --output text)

    echo "Updated and published version $NEW_VERSION"

    # Create an alias for the new version
    ALIAS_NAME="alias_$i"
    aws lambda create-alias --function-name $FUNCTION_NAME --name $ALIAS_NAME --function-version $NEW_VERSION --region $AWS_REGION --profile automation

    # Check if the alias is created successfully
    if [ $? -ne 0 ]; then
        echo "Failed to create alias $ALIAS_NAME."
    else
        echo "Created alias $ALIAS_NAME for version $NEW_VERSION."
    fi

    # Remove the temporary directory
    rm -r $TEMP_DIR
done

echo "Script completed!"
