#!/bin/bash

# Define the output file with a .txt extension
OUTPUT_FILE="all_in_one.txt"

# Header for the output file to explain its purpose
echo "// Concatenated Go files for review" > $OUTPUT_FILE
echo "// This file is generated for readability and may not compile as is." >> $OUTPUT_FILE
echo "// Although this is Go code, it's saved in a .txt file for easier sharing and review." >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Since we're now appending to the file, ensure we don't duplicate the header by removing it if it exists
# Adjust the line number '1,4d' based on the number of header lines
sed -i '1,4d' $OUTPUT_FILE

# Find all .go files and concatenate them into the output file
# Excludes the vendor directory and the output file itself
find . -type f -name '*.go' ! -path "./vendor/*" ! -name $OUTPUT_FILE | while read filename; do
    echo "// File: $filename" >> $OUTPUT_FILE
    cat "$filename" >> $OUTPUT_FILE
    echo "" >> $OUTPUT_FILE # Add a newline for separation
    echo "// End of $filename" >> $OUTPUT_FILE
    echo "" >> $OUTPUT_FILE # Extra newline for readability
done

# Display a message with the result
echo "All Go files have been concatenated into $OUTPUT_FILE"