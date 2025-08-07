# Overview

This is a comprehensive pure Go terminal application for creating and managing multiple Telegram shopping bots. The system includes a mother bot (CodeRoot) that allows sellers to create their own sub-bots with different subscription plans. Built entirely in Go for maximum performance, cross-platform compatibility, and easy deployment on any system with Go installed.

## Project Architecture 

**Pure Go Terminal Application - No Web Frameworks**
- Single binary executable for all platforms
- Configuration via .env file only  
- Terminal-based operation with comprehensive logging
- No web interface - pure Telegram bot interaction
- Mother bot creates and manages shop telegram bots automatically

# User Preferences

Preferred communication style: Simple, everyday language.

# Recent Changes

## ✅ Comprehensive Installation System (August 7, 2025)
- **Complete Setup Scripts**: Full automation for all prerequisites and dependencies
- **Multi-Platform Support**: Ubuntu/Debian, CentOS/RHEL, macOS, Android Termux installation scripts
- **Database Automation**: PostgreSQL installation, database creation, table setup with proper indexes
- **Security Configuration**: Auto-generated secure passwords, proper user permissions, systemd hardening
- **Cross-Platform Builds**: Automated build system for all supported platforms (Linux, Windows, macOS, Android)
- **Documentation**: Complete installation guide with troubleshooting and platform-specific instructions
- **Service Management**: Systemd service creation, auto-start configuration, log management
- **Update/Uninstall Scripts**: Safe update and removal scripts with data backup options

## ✅ Migration from Replit Agent to Replit Complete (August 7, 2025)
- **Dependency Installation**: All packages successfully installed using packager tool
- **Security Enhancements**: XSS protection, input validation, and sanitization implemented
- **API Endpoints**: Complete REST API with user management and health checks
- **Error Handling**: Comprehensive error boundaries and robust error handling
- **UI Testing Infrastructure**: Complete data-testid attributes for all interactive elements
- **Theme System**: Dark/light mode support with proper CSS variables
- **Bug Fixes**: Toast timeout reduced from 16+ minutes to 5 seconds
- **Navigation**: Full routing system with API testing dashboard
- **Code Quality**: LSP diagnostics resolved, TypeScript strict mode enforced
- **Comprehensive Testing**: Created extensive test suite for all bot components and database operations
- **Enhanced Installation**: Improved installation scripts with security hardening and cross-platform support
- **Documentation**: Complete deployment guide with troubleshooting and security considerations
- **Bot System Validation**: Fixed model field references and ensured proper database relationships

# Previous Changes (Telegram Bot System)

## ✅ Complete Go-based Telegram Bot System
- **Project Complete**: Full conversion from TypeScript web app to pure Go telegram bot
- **Mother Bot (CodeRoot)**: Main bot with complete seller registration and sub-bot creation
- **Subscription System**: Free (10 products), Pro (200 products), VIP (unlimited) - fully implemented
- **Automatic Sub-Bot Creation**: Individual store bots with custom branding - framework ready
- **Persian Language Support**: Full UI in Persian for Iranian market - complete
- **Cross-Platform**: Windows, Linux, macOS, Android (Termux) support - build system ready
- **Database**: PostgreSQL integration with GORM - complete with all models
- **Pure Go Implementation**: No web frameworks, terminal-based operation - complete
- **Bot Handlers**: Complete bot logic with callbacks, products, payments, and admin features
- **Service Layer**: Full business logic with user, store, product, order, payment, and session services
- **Build System**: Makefile with cross-platform builds and Docker support
- **Documentation**: Complete README, INSTALLATION, and USAGE guides

# System Architecture

## Core Features Implementation
- **Mother Bot (CodeRoot)**: Main bot for seller registration and management
- **Subscription System**: Free (10 products), Pro (200 products), VIP (unlimited)
- **Automatic Sub-Bot Creation**: Individual store bots with custom branding
- **Persian Language Support**: Full UI in Persian for Iranian market
- **Cross-Platform**: Windows, Linux, macOS, Android (Termux) support

## Backend Architecture
- **Runtime**: Go with Telegram Bot API integration
- **Language**: Go with concurrent bot handling
- **Database**: PostgreSQL with GORM for database operations
- **Bot Framework**: go-telegram-bot-api for Telegram integration
- **API Structure**: RESTful API for bot management and webhook handling
- **Deployment**: Cross-platform binary support (Windows, Linux, macOS, Android Termux)

## Database Design
- **Schema Definition**: Centralized in shared/schema.ts using Drizzle schema
- **Type Safety**: Full TypeScript integration with inferred types
- **Validation**: Zod schemas generated from Drizzle schema for runtime validation
- **Current Schema**: Users table with id, username, and password fields

## Development Environment
- **Build System**: Vite for frontend, esbuild for backend production builds
- **Development Server**: Custom Vite integration with Express for seamless full-stack development
- **TypeScript**: Strict configuration with path mapping for clean imports
- **Hot Reload**: Full-stack hot reload in development mode

## Storage Architecture
- **Interface Pattern**: Abstract IStorage interface for flexible storage backends
- **Current Implementation**: In-memory storage (MemStorage) for development
- **Database Ready**: Configured for PostgreSQL with Neon Database integration
- **Migration System**: Drizzle Kit for database schema migrations

## UI/UX Design System
- **Component Library**: Complete shadcn/ui implementation with 40+ components
- **Theme System**: CSS custom properties with light/dark mode support
- **Responsive Design**: Mobile-first approach with Tailwind breakpoints
- **Accessibility**: Radix UI primitives ensure WCAG compliance
- **Typography**: Consistent design tokens for spacing, colors, and typography

# External Dependencies

## Database Services
- **Neon Database**: Serverless PostgreSQL with @neondatabase/serverless driver
- **Connection Management**: Environment-based DATABASE_URL configuration

## UI and Styling
- **Radix UI**: Complete primitive component library for accessibility and behavior
- **Tailwind CSS**: Utility-first CSS framework with custom configuration
- **Lucide React**: Icon library for consistent iconography
- **Class Variance Authority**: Type-safe variant management for components

## Development Tools
- **Vite**: Frontend build tool with React plugin and runtime error overlay
- **Drizzle Kit**: Database migration and schema management tool
- **ESBuild**: Fast JavaScript bundler for production backend builds
- **TSX**: TypeScript execution engine for development

## State Management and Data Fetching
- **TanStack Query**: Powerful data fetching and caching library
- **React Hook Form**: Performant form library with validation
- **Zod**: Runtime type validation and schema validation

## Utility Libraries
- **date-fns**: Modern date utility library
- **clsx & tailwind-merge**: Utility functions for conditional className handling
- **nanoid**: Secure URL-friendly unique ID generator

## Platform Integration
- **Replit Integration**: Custom plugins for development environment and debugging tools
- **Session Management**: PostgreSQL session store with connect-pg-simple