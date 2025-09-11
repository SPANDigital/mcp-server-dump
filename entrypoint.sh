#!/bin/sh
set -e

# Function to log messages
log() {
    echo "::notice::$1"
}

# Function to set output
set_output() {
    echo "$1=$2" >> "$GITHUB_OUTPUT"
}

# Function to handle errors
error_exit() {
    echo "::error::$1"
    exit 1
}

log "Starting MCP Server Dump Action"

# Build the command
CMD="mcp-server-dump"

# Add transport if specified
if [ -n "$INPUT_TRANSPORT" ] && [ "$INPUT_TRANSPORT" != "stdio" ]; then
    CMD="$CMD -t $INPUT_TRANSPORT"
fi

# Add endpoint if specified
if [ -n "$INPUT_ENDPOINT" ]; then
    CMD="$CMD --endpoint $INPUT_ENDPOINT"
fi

# Add headers if specified
if [ -n "$INPUT_HEADERS" ]; then
    # Split headers by comma and add each one
    echo "$INPUT_HEADERS" | tr ',' '\n' | while read -r header; do
        if [ -n "$header" ]; then
            CMD="$CMD -H \"$header\""
        fi
    done
fi

# Add format if not default
if [ "$INPUT_FORMAT" != "markdown" ]; then
    CMD="$CMD -f $INPUT_FORMAT"
fi

# Add output file if specified
if [ -n "$INPUT_OUTPUT_FILE" ]; then
    CMD="$CMD -o $INPUT_OUTPUT_FILE"
    OUTPUT_FILE="$INPUT_OUTPUT_FILE"
else
    # Generate default output filename based on format
    case "$INPUT_FORMAT" in
        "html")
            OUTPUT_FILE="mcp-server-dump.html"
            CMD="$CMD -o $OUTPUT_FILE"
            ;;
        "json")
            OUTPUT_FILE="mcp-server-dump.json"
            CMD="$CMD -o $OUTPUT_FILE"
            ;;
        "pdf")
            error_exit "PDF format requires output-file to be specified"
            ;;
        *)
            OUTPUT_FILE="mcp-server-dump.md"
            CMD="$CMD -o $OUTPUT_FILE"
            ;;
    esac
fi

# Add no-toc flag if specified
if [ "$INPUT_NO_TOC" = "true" ]; then
    CMD="$CMD --no-toc"
fi

# Add frontmatter if specified
if [ -n "$INPUT_FRONTMATTER" ]; then
    CMD="$CMD --frontmatter $INPUT_FRONTMATTER"
fi

# Add timeout if not default
if [ "$INPUT_TIMEOUT" != "30" ]; then
    CMD="$CMD --timeout ${INPUT_TIMEOUT}s"
fi

# Add verbose flag if specified
if [ "$INPUT_VERBOSE" = "true" ]; then
    CMD="$CMD -v"
fi

# Add server command (this should be the last argument)
if [ -n "$INPUT_SERVER_COMMAND" ]; then
    CMD="$CMD $INPUT_SERVER_COMMAND"
fi

log "Executing: $CMD"

# Execute the command
if eval "$CMD"; then
    log "MCP Server Dump completed successfully"
    
    # Set outputs
    set_output "output-file" "$OUTPUT_FILE"
    
    # If the output file exists, try to extract server info for the output
    if [ -f "$OUTPUT_FILE" ] && [ "$INPUT_FORMAT" = "json" ]; then
        SERVER_INFO=$(cat "$OUTPUT_FILE")
        set_output "server-info" "$SERVER_INFO"
    elif [ -f "$OUTPUT_FILE" ]; then
        # For non-JSON formats, create a simple server info
        SERVER_INFO="{\"output_file\":\"$OUTPUT_FILE\",\"format\":\"$INPUT_FORMAT\"}"
        set_output "server-info" "$SERVER_INFO"
    fi
    
    log "Output file: $OUTPUT_FILE"
else
    error_exit "MCP Server Dump failed"
fi