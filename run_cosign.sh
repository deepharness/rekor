#!/bin/bash

# Set the COSIGN_PASSWORD environment variable
export COSIGN_PASSWORD=1234
echo "COSIGN_PASSWORD set to 1234"

# Prompt the user for the total number of iterations
read -p "Enter the total number of times to run the command: " total
# Prompt the user for the delay in milliseconds between executions
read -p "Enter the delay in milliseconds between each execution: " delay_ms

# Check if the inputs are valid numbers
if ! [[ "$total" =~ ^[0-9]+$ ]] || ! [[ "$delay_ms" =~ ^[0-9]+$ ]]; then
  echo "Error: Please enter valid numeric values for both the number of iterations and delay."
  exit 1
fi

# Convert milliseconds to seconds (fractional)
delay_seconds=$(echo "scale=3; $delay_ms / 1000" | bc)

# Function to execute a single instance and print the tlog index
run_command() {
  output=$(cosign sign-blob --y README.md --key cosign.key --bundle cosign.bundle --rekor-url http://localhost:8080/ 2>&1)
  index=$(echo "$output" | grep "tlog entry created with index:" | sed -n 's/.*index: \([0-9]*\).*/\1/p')
  if [[ -n "$index" ]]; then
    echo "tlog entry created with index: $index"
  fi
  sleep "$delay_seconds"
}

# Batch size
batch_size=100

# Divide the total number of iterations into batches
for ((start=0; start<total; start+=batch_size)); do
  echo "Executing batch starting from $((start + 1))"
  
  # Start time for the batch
  start_time=$(date +%s.%N)
  
  # Spawn parallel jobs for the current batch
  for ((i=0; i<batch_size && start+i<total; i++)); do
    run_command &
  done

  # Wait for all parallel jobs in the current batch to complete
  wait

  # End time for the batch
  end_time=$(date +%s.%N)

  # Calculate and display the total time taken for the batch
  total_time=$(echo "$end_time - $start_time" | bc)
  echo "Batch completed. Time taken: ${total_time} seconds."
done

echo "Completed $total executions."
