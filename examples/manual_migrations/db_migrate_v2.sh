#!/bin/bash
# Direct database migration runner
# This script handles all migration operations directly with the database

# Default parameters
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="MyPassword_123"
DB_NAME="gra"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display usage
usage() {
  echo -e "${BLUE}GRA Database Migration Tool${NC}"
  echo "Usage: $0 <command> [options]"
  echo ""
  echo "Commands:"
  echo "  up                Apply all pending migrations"
  echo "  down              Roll back the last migration"
  echo "  status            Show migration status"
  echo "  test              Test database connection"
  echo ""
  echo "Options:"
  echo "  --host STRING     Database host (default: $DB_HOST)"
  echo "  --port STRING     Database port (default: $DB_PORT)"
  echo "  --user STRING     Database user (default: $DB_USER)"
  echo "  --password STRING Database password (default: $DB_PASSWORD)"
  echo "  --dbname STRING   Database name (default: $DB_NAME)"
  echo "  --verbose         Show verbose output"
  exit 1
}

# Parse command-line arguments
COMMAND=""
VERBOSE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    up|down|status|test|to)
      COMMAND=$1
      shift
      ;;
    --host)
      DB_HOST="$2"
      shift 2
      ;;
    --port)
      DB_PORT="$2"
      shift 2
      ;;
    --user)
      DB_USER="$2"
      shift 2
      ;;
    --password)
      DB_PASSWORD="$2"
      shift 2
      ;;
    --dbname)
      DB_NAME="$2"
      shift 2
      ;;
    --verbose)
      VERBOSE=true
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo -e "${RED}Error: Unknown option: $1${NC}"
      usage
      ;;
  esac
done

# If no command specified, show usage
if [[ -z "$COMMAND" ]]; then
  usage
fi

# Create connection string
DB_URI="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

# Function to test database connection
test_connection() {
  echo -e "${BLUE}Testing database connection...${NC}"
  if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: PostgreSQL client not found.${NC}"
    echo "Please install PostgreSQL client tools."
    exit 1
  fi
  
  if psql "$DB_URI" -c '\conninfo' &> /dev/null; then
    echo -e "${GREEN}Connection successful!${NC}"
    return 0
  else
    echo -e "${RED}Connection failed!${NC}"
    echo "Make sure PostgreSQL is running and the connection parameters are correct:"
    echo "  Host:     $DB_HOST"
    echo "  Port:     $DB_PORT"
    echo "  User:     $DB_USER"
    echo "  Database: $DB_NAME"
    return 1
  fi
}



# Function to ensure Go modules are properly set up
ensure_go_modules() {
  if [ "$VERBOSE" = true ]; then
    echo -e "${BLUE}Ensuring Go modules...${NC}"
  fi
  
  # Make sure we're using Go modules
  export GO111MODULE=on
  
  # Check if github.com/lib/pq is installed
  if ! go list -m github.com/lib/pq &> /dev/null; then
    echo -e "${YELLOW}Installing PostgreSQL driver...${NC}"
    go get github.com/lib/pq
  fi
}

# Function to execute migration commands with useful feedback
execute_migration() {
  local action=$1
  local args=${@:2}
  
  ensure_go_modules
  
  echo -e "${BLUE}Executing migration: $action${NC}"
  
  # Ensure the direct runner is built
  if [ ! -f "../../tools/migration/direct_runner" ]; then
    echo -e "${YELLOW}Building direct runner...${NC}"
    go build -o ../../tools/migration/direct_runner ../../tools/migration/direct_runner.go
  fi
  
  # Use direct runner with database connection for real migrations
  local cmd="../../tools/migration/direct_runner --conn '$DB_URI' --$action"
  
  if [ "$VERBOSE" = true ]; then
    echo -e "${YELLOW}Command: $cmd${NC}"
  fi
  
  # Execute the command
  if eval "$cmd"; then
    echo -e "${GREEN}Migration completed successfully.${NC}"
    return 0
  else
    echo -e "${RED}Migration failed!${NC}"
    return 1
  fi
}

# Main execution
echo -e "${BLUE}GRA Database Migration Tool${NC}"

case $COMMAND in
  test)
    test_connection
    ;;
    
  status)
    # Test connection first
    if ! test_connection; then
      exit 1
    fi
    
    echo -e "${BLUE}\nMigration Status:${NC}"
    
    # Use direct_runner for status
    if [ ! -f "../../tools/migration/direct_runner" ]; then
      echo -e "${YELLOW}Building direct_runner...${NC}"
      go build -o ../../tools/migration/direct_runner ../../tools/migration/direct_runner.go
    fi
    ../../tools/migration/direct_runner --conn "$DB_URI" --status
    ;;
    
  up)
    # Test connection first
    if ! test_connection; then
      exit 1
    fi
    
    # Execute up command
    echo -e "${BLUE}\nApplying pending migrations...${NC}"
    execute_migration "up" "--verbose"
    ;;
    
  down)
    # Test connection first
    if ! test_connection; then
      exit 1
    fi
    
    # Execute down command
    echo -e "${BLUE}\nRolling back last migration...${NC}"
    execute_migration "down" "--verbose"
    ;;
    
  to)
    echo -e "${RED}Error: 'to' command is not supported in this version${NC}"
    echo -e "${YELLOW}Use 'up' to apply all pending migrations or 'down' to rollback${NC}"
    exit 1
    ;;
    
  *)
    echo -e "${RED}Error: Unknown command: $COMMAND${NC}"
    usage
    ;;
esac
