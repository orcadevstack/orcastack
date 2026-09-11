#!/bin/bash

# Docker Build and Deployment Script
# This script helps you build and manage Docker containers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}$1${NC}"
    echo -e "${GREEN}========================================${NC}"
}

print_error() {
    echo -e "${RED} Error: $1${NC}"
}

print_success() {
    echo -e "${GREEN} $1${NC}"
}

print_info() {
    echo -e "${YELLOW}  $1${NC}"
}

# Check if docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    print_success "Docker is installed"
}

check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    print_success "Docker Compose is installed"
}

# Build functions
build_all() {
    print_header "Building all services"
    docker-compose build
    print_success "All services built successfully"
}

build_backend() {
    print_header "Building backend service"
    docker-compose build backend
    print_success "Backend built successfully"
}

build_frontend() {
    print_header "Building frontend service"
    docker-compose build frontend
    print_success "Frontend built successfully"
}

# Start functions
start_all() {
    print_header "Starting all services"
    docker-compose up -d
    print_success "All services started"
    print_info "Backend: http://localhost:8000"
    print_info "Frontend: http://localhost:3000"
    print_info "Database: localhost:5432"
}

start_development() {
    print_header "Starting in development mode"
    docker-compose up
}

stop_all() {
    print_header "Stopping all services"
    docker-compose down
    print_success "All services stopped"
}

restart_all() {
    print_header "Restarting all services"
    docker-compose restart
    print_success "All services restarted"
}

# Log functions
show_logs() {
    print_header "Showing logs for all services"
    docker-compose logs -f
}

show_logs_backend() {
    print_header "Showing logs for backend"
    docker-compose logs -f backend
}

show_logs_frontend() {
    print_header "Showing logs for frontend"
    docker-compose logs -f frontend
}

show_logs_db() {
    print_header "Showing logs for database"
    docker-compose logs -f db
}

# Database functions
migrate() {
    print_header "Running database migrations"
    docker-compose exec backend python manage.py migrate
    print_success "Migrations completed"
}

add_sample_data() {
    print_header "Adding sample data"
    docker-compose exec backend python manage.py add_sample_discussions
    print_success "Sample data added"
}

create_superuser() {
    print_header "Creating superuser"
    docker-compose exec backend python manage.py createsuperuser
}

backup_database() {
    print_header "Backing up database"
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    BACKUP_FILE="backup_${TIMESTAMP}.sql"
    docker-compose exec db pg_dump -U postgres personal_brand_hub > "$BACKUP_FILE"
    print_success "Database backed up to $BACKUP_FILE"
}

# Status functions
show_status() {
    print_header "Container Status"
    docker-compose ps
}

show_stats() {
    print_header "Container Statistics"
    docker stats
}

# Clean functions
clean_all() {
    print_header "Cleaning up all Docker resources"
    read -p "This will remove all containers, volumes, and networks. Are you sure? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose down -v
        print_success "Cleanup completed"
    else
        print_info "Cleanup cancelled"
    fi
}

prune_system() {
    print_header "Pruning unused Docker resources"
    docker system prune -f
    print_success "Prune completed"
}

# Help function
show_help() {
    echo "Docker Build and Deployment Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build-all              Build all services"
    echo "  build-backend          Build backend service"
    echo "  build-frontend         Build frontend service"
    echo "  start                  Start all services in background"
    echo "  start-dev              Start all services with logs in foreground"
    echo "  stop                   Stop all services"
    echo "  restart                Restart all services"
    echo "  status                 Show container status"
    echo "  logs                   Show logs for all services"
    echo "  logs-backend           Show logs for backend"
    echo "  logs-frontend          Show logs for frontend"
    echo "  logs-db                Show logs for database"
    echo "  migrate                Run database migrations"
    echo "  add-sample             Add sample data to database"
    echo "  create-user            Create a superuser"
    echo "  backup                 Backup database"
    echo "  stats                  Show container statistics"
    echo "  clean                  Remove all containers and volumes"
    echo "  prune                  Prune unused Docker resources"
    echo "  help                   Show this help message"
    echo ""
}

# Main script
main() {
    case "${1:-help}" in
        build-all)
            check_docker && check_docker_compose && build_all
            ;;
        build-backend)
            check_docker && check_docker_compose && build_backend
            ;;
        build-frontend)
            check_docker && check_docker_compose && build_frontend
            ;;
        start)
            check_docker && check_docker_compose && start_all
            ;;
        start-dev)
            check_docker && check_docker_compose && start_development
            ;;
        stop)
            check_docker && check_docker_compose && stop_all
            ;;
        restart)
            check_docker && check_docker_compose && restart_all
            ;;
        status)
            check_docker && check_docker_compose && show_status
            ;;
        logs)
            check_docker && check_docker_compose && show_logs
            ;;
        logs-backend)
            check_docker && check_docker_compose && show_logs_backend
            ;;
        logs-frontend)
            check_docker && check_docker_compose && show_logs_frontend
            ;;
        logs-db)
            check_docker && check_docker_compose && show_logs_db
            ;;
        migrate)
            check_docker && check_docker_compose && migrate
            ;;
        add-sample)
            check_docker && check_docker_compose && add_sample_data
            ;;
        create-user)
            check_docker && check_docker_compose && create_superuser
            ;;
        backup)
            check_docker && check_docker_compose && backup_database
            ;;
        stats)
            check_docker && check_docker_compose && show_stats
            ;;
        clean)
            check_docker && check_docker_compose && clean_all
            ;;
        prune)
            check_docker && check_docker_compose && prune_system
            ;;
        help)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
