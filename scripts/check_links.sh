#!/bin/bash
# Script to check for broken links in the documentation

echo "Checking for broken links in the GRA Framework documentation..."
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

DOCS_DIR="docs"
BROKEN_LINKS=0
TOTAL_LINKS=0
INTERNAL_LINKS=0
EXTERNAL_LINKS=0

# Function to extract links from a markdown file
extract_links() {
    local file=$1
    # Extract markdown links [text](url) and HTML links <a href="url">
    grep -o -E '\[.+?\]\(.+?\)|\<a href="[^"]+' "$file" | \
    sed -E 's/\[.+\]\((.+)\)/\1/g' | \
    sed -E 's/<a href="([^"]+).*/\1/g'
}

# Check if a link is valid
check_link() {
    local link=$1
    local source_file=$2
    local is_valid=true
    local reason=""
    
    TOTAL_LINKS=$((TOTAL_LINKS + 1))
    
    # Skip fragments in the same file
    if [[ "$link" =~ ^# ]]; then
        link="$source_file$link"
    fi

    # External links (http/https)
    if [[ "$link" =~ ^https?:// ]]; then
        EXTERNAL_LINKS=$((EXTERNAL_LINKS + 1))
        echo -n "."
        # Don't actually check external links to avoid rate limiting
        return 0
    else
        INTERNAL_LINKS=$((INTERNAL_LINKS + 1))
        
        # Handle absolute internal links
        if [[ "$link" = /* ]]; then
            link="$DOCS_DIR$link"
        fi
        
        # Handle relative links
        if [[ ! "$link" =~ ^/ && ! "$link" =~ ^# ]]; then
            # Get directory of source file
            local dir=$(dirname "$source_file")
            link="$dir/$link"
        fi
        
        # Remove query string and fragment
        link=$(echo "$link" | sed 's/[?#].*$//')
        
        # Remove trailing slash
        link=$(echo "$link" | sed 's/\/$//')
        
        # Check if file exists
        if [[ "$link" && ! -e "$link" ]]; then
            # Check if it's a directory by adding /README.md
            if [[ ! -e "$link/README.md" && ! -e "$link/index.html" ]]; then
                is_valid=false
                reason="File not found"
            fi
        fi
    fi
    
    if [[ "$is_valid" = false ]]; then
        echo -e "\n${RED}[ERROR]${NC} Broken link in $source_file: $link - $reason"
        BROKEN_LINKS=$((BROKEN_LINKS + 1))
        return 1
    fi
    
    echo -n "."
    return 0
}

# Find all markdown files
echo "Scanning markdown files..."
find "$DOCS_DIR" -name "*.md" | while read -r file; do
    echo -e "\nChecking links in ${YELLOW}$file${NC}"
    
    # Extract and check each link
    extract_links "$file" | while read -r link; do
        check_link "$link" "$file"
    done
done

echo -e "\n\n${GREEN}Link check completed!${NC}"
echo "Total links checked: $TOTAL_LINKS (Internal: $INTERNAL_LINKS, External: $EXTERNAL_LINKS)"

if [ "$BROKEN_LINKS" -gt 0 ]; then
    echo -e "${RED}Found $BROKEN_LINKS broken links!${NC}"
    exit 1
else
    echo -e "${GREEN}No broken links found!${NC}"
    exit 0
fi
