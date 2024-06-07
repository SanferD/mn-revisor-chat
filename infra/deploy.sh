#!/bin/bash

# Read the nonce value from config.yaml
echo "extracting nonce"
nonce=$(grep "nonce" config.yaml | awk -F': ' '{print $2}' | tr -d '"')
echo "found nonce \"$nonce\""

# Run cdk list and filter stack names based on the nonce
stack_names=$(cdk list | grep "\-stack\-$nonce")

# Check if stack names were found
if [ -z "$stack_names" ]; then
  echo "No stack names found with nonce $nonce. Exiting."
  exit 1
fi
echo "found stacks:\n\n$stack_names"
echo

echo "Starting deployment process..."

# Loop over each stack name and deploy
for stack in $stack_names; do
  echo "Deploying $stack..."
  cdk deploy --require-approval never $stack
  if [ $? -ne 0 ]; then
    echo "Deployment of $stack failed. Exiting."
    exit 1
  fi
done
