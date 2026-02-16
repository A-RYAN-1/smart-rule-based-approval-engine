# Smart Rule-Based Approval Engine

A comprehensive, **production-ready** approval automation system that streamlines request management with intelligent rule-based decision making. This platform handles multiple request types (Leave, Expense, Discount) with real-time processing, automated approvals/rejections, and a powerful admin dashboard.

**Live Demo:** [smart-rule-based-approval-engine.vercel.app](https://smart-rule-based-approval-engine.vercel.app)

---

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

---

## 🎯 Overview

**Approval Genie** automates workflow approvals for modern enterprises with intelligent rule-based decision making. Instead of manual approvals, the system automatically processes requests based on custom rules, saving time and improving consistency.

---

## ✨ Features

### 📝 Request Management
- **Leave Requests** - Auto-approve/reject based on leave balance, grade, and custom rules
- **Expense Reports** - Intelligent expense categorization and approval routing
- **Discount Requests** - Domain-specific discount approvals with business logic
- **Auto Rejection** - Automated rejection of out-of-policy requests
- **Request History** - Complete audit trail of all requests

### 👤 Employee Features
- Apply for leaves, expenses, and discounts
- Track request status in real-time
- View approval history and feedback
- Download reports and documentation
- Responsive mobile-friendly interface

### 👨‍💼 Manager/Approver Features
- View pending requests dashboard
- Approve/reject with custom comments
- Bulk operations on requests
- Filter by request type, date range, requester
- Export data for reporting

### 🔧 Admin Dashboard
- **Rules Engine** - Create/edit/delete approval rules
- **Holiday Management** - Configure company holidays
- **Advanced Reports** - Status distribution, request analytics
- **User Management** - Add/manage employees and approvers
- **System Logs** - Monitor system activities
- **Configurations** - Grade levels, approval thresholds

### 🤖 Intelligent Features
- **Smart Routing** - Automatically assign requests to right approvers
- **Rule Conditions** - Complex rule evaluation (grade, amount, department)
- **Auto Decisions** - Instant approval/rejection based on rules
- **Balance Tracking** - Real-time leave balance calculations
- **Cron Jobs** - Scheduled auto-rejection of expired requests

---

## 🛠️ Tech Stack

### Frontend
- **Framework:** React 18 with TypeScript
- **Styling:** Tailwind CSS + Shadcn UI
- **State Management:** TanStack Query + Zustand
- **Build Tool:** Vite
- **Testing:** Vitest with React Testing Library
- **HTTP Client:** Axios
- **Charts:** Recharts

### Backend
- **Language:** Go 1.25+
- **Framework:** Gin Web Framework
- **Database:** PostgreSQL (via Supabase)
- **Database Driver:** pgx
- **Migrations:** Custom migration system
- **Authentication:** JWT tokens
- **Validation:** Go struct tags

### Infrastructure
- **Frontend Hosting:** Vercel
- **Backend Hosting:** Render
- **Database:** Supabase PostgreSQL
- **Version Control:** Git

---

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ and npm/pnpm
- Go 1.25+
- PostgreSQL 14+ (or Supabase account)
- Git

### Local Development Setup

#### 1. Clone the Repository
```bash
git clone https://github.com/A-RYAN-1/smart-rule-based-approval-engine.git
cd smart-rule-based-approval-engine
```

#### 2. Backend Setup

```bash
cd backend

# Install dependencies (if using go modules)
go mod download

# Set up environment variables
cp .env.example .env

# Configure your database connection
# Edit .env and add:
# DATABASE_URL=postgresql://user:password@localhost:5432/approval_engine
# JWT_SECRET=your_secret_key
# GIN_MODE=debug

# Run migrations
go run ./cmd/server migrate up

# Start the server
go run ./cmd/server
# Server runs on http://localhost:8080
```

#### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Set up environment variables
cp .env.example .env.local

# Configure API endpoint
# VITE_API_BASE_URL=http://localhost:8080

# Start development server
npm run dev
# Frontend runs on http://localhost:5173
```

#### 4. Access the Application
- **Frontend:** http://localhost:5173
- **Backend API:** http://localhost:8080

---

## 📁 Project Structure

```
smart-rule-based-approval-engine/
├── backend/                          # Go backend application
│   ├── cmd/server/                   # Application entry point
│   ├── app/                          # Business logic
│   │   ├── auth/                     # Authentication handlers
│   │   ├── leave_service/            # Leave request logic
│   │   ├── expense_service/          # Expense request logic
│   │   ├── domain_service/           # Domain/discount logic
│   │   ├── rules/                    # Rules engine
│   │   ├── holidays/                 # Holiday management
│   │   ├── reports/                  # Reporting logic
│   │   └── my_requests/              # User request dashboard
│   ├── models/                       # Data models
│   ├── repositories/                 # Database access layer
│   │   ├── aggregated_repository.go  # Unified data access
│   │   ├── balance_repository.go     # Leave balance queries
│   │   ├── rule_repository.go        # Rules queries
│   │   └── ...                       # Other repositories
│   ├── migrations/                   # SQL migration files
│   │   ├── 1770125704_add_users_table.up.sql
│   │   └── 1770125704_add_users_table.down.sql
│   ├── config/                       # Configuration management
│   ├── database/                     # Database connection
│   ├── pkg/                          # Shared packages
│   │   ├── middleware/               # HTTP middleware
│   │   ├── utils/                    # Utility functions
│   │   ├── response/                 # Response formatting
│   │   └── apperrors/                # Error handling
│   ├── routes/                       # API route definitions
│   ├── interfaces/                   # Service interfaces
│   ├── main.go                       # Server initialization
│   ├── go.mod                        # Go module file
│   └── Makefile                      # Build commands
│
├── frontend/                         # React/TypeScript frontend
│   ├── src/
│   │   ├── components/               # Reusable React components
│   │   ├── pages/                    # Page components
│   │   ├── hooks/                    # Custom React hooks
│   │   ├── services/                 # API client services
│   │   ├── types/                    # TypeScript type definitions
│   │   ├── utils/                    # Utility functions
│   │   ├── stores/                   # Zustand state stores
│   │   ├── App.tsx                   # Root component
│   │   └── main.tsx                  # React entry point
│   ├── public/                       # Static assets
│   ├── index.html                    # HTML template
│   ├── package.json                  # Dependencies
│   ├── vite.config.ts                # Vite configuration
│   ├── tsconfig.json                 # TypeScript configuration
│   └── tailwind.config.ts            # Tailwind CSS configuration
│
└── docs/                             # Documentation files
    ├── DEPLOYMENT_*.md               # Deployment guides
    ├── DATABASE_SCHEMA_OVERVIEW.md   # Database schema
    └── MIGRATION_GUIDE.md            # Migration documentation
```

---

## 🗄️ Database Schema

The application uses PostgreSQL with the following main tables:

| Table | Purpose |
|-------|---------|
| `users` | Employee and approver information |
| `leave_requests` | Leave application records |
| `leave_balances` | Current leave balance per user |
| `expense_requests` | Expense report submissions |
| `discount_requests` | Discount application requests |
| `rules` | Approval rule definitions |
| `holidays` | Company holiday calendar |
| `reports` | Generated reports and analytics |
| `audit_logs` | System activity logging |
| `schema_migrations` | Migration tracking |

---

## 📦 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Get current user info

### Leave Requests
- `POST /api/leave/apply` - Submit leave request
- `GET /api/leaves/my` - Get user's leave requests
- `GET /api/leaves/pending` - Get pending approvals
- `POST /api/leaves/:id/approve` - Approve leave request
- `POST /api/leaves/:id/reject` - Reject leave request
- `POST /api/leaves/:id/cancel` - Cancel leave request

### Expense Requests
- `POST /api/expenses/request` - Submit expense request
- `GET /api/expenses/my` - Get user's expenses
- `GET /api/expenses/pending` - Get pending approvals
- `POST /api/expenses/:id/approve` - Approve expense
- `POST /api/expenses/:id/reject` - Reject expense
- `POST /api/expenses/:id/cancel` - Cancel expense

### Discount Requests
- `POST /api/discounts/request` - Submit discount request
- `GET /api/discounts/my` - Get user's discounts
- `GET /api/discounts/pending` - Get pending approvals
- `POST /api/discounts/:id/approve` - Approve discount
- `POST /api/discounts/:id/reject` - Reject discount

### Admin Operations
- `POST /api/admin/rules` - Create approval rule
- `GET /api/admin/rules` - List all rules
- `PUT /api/admin/rules/:id` - Update rule
- `DELETE /api/admin/rules/:id` - Delete rule
- `POST /api/admin/holidays` - Add holiday
- `GET /api/admin/holidays` - List holidays
- `DELETE /api/admin/holidays/:id` - Delete holiday
- `GET /api/admin/reports/*` - Generate reports

See [API.txt](./backend/API.txt) or [swagger.yaml](./backend/api/swagger.yaml) for complete documentation.

---

## 🚀 Deployment

### Quick Deployment Guide

This project is designed for **free deployment** on:

1. **Frontend:** Vercel (Global CDN)
2. **Backend:** Render (Auto-scaling)
3. **Database:** Supabase (PostgreSQL)
4. **Monitoring:** UptimeRobot (24/7)

### Prerequisites for Deployment

You'll need accounts for:
- **GitHub** - For source code and CI/CD
- **Supabase** - PostgreSQL database
- **Render** - Backend hosting
- **Vercel** - Frontend hosting
- **UptimeRobot** - Monitoring (optional)

### Deployment Steps

#### 1. Create Supabase Database
- Go to https://supabase.com
- Create a new project
- Get connection string from Project Settings
- **Important:** The migrations will run automatically on backend startup

#### 2. Configure Backend Environment (Render)
Set these environment variables on Render:
```
DATABASE_URL=postgresql://[user]:[password]@[host]/[database]
JWT_SECRET=your_secure_secret_key
GIN_MODE=release
```

#### 3. Deploy Backend
```bash
cd backend
git push origin main
# Render will automatically detect Go project and deploy
```

#### 4. Configure Frontend Environment (Vercel)
Set environment variable:
```
VITE_API_BASE_URL=https://your-backend-url
```

#### 5. Deploy Frontend
```bash
cd frontend
git push origin main
# Vercel will automatically detect and deploy
```

#### 6. Configure GitHub Secrets
Add to GitHub Actions secrets:
```
SUPABASE_URL=your_supabase_url
SUPABASE_KEY=your_supabase_key
DATABASE_URL=your_full_database_url
```

For detailed deployment instructions, see:
- [DEPLOYMENT_QUICK_START.md](./docs/DEPLOYMENT_QUICK_START.md)
- [DEPLOYMENT_CREDENTIALS_NEEDED.md](./docs/DEPLOYMENT_CREDENTIALS_NEEDED.md)

---

## ❌ Troubleshooting

### Issue: "Migrations not found" or "No tables in database"

**Symptoms:**
- Supabase shows empty database
- Backend logs show: `files []`
- 404 errors on API endpoints

**Solution:**
The migrations must be in the correct directory and discoverable:

1. Ensure migrations are in `backend/migrations/` directory
2. Check that `migrate.go` can find the migrations
3. The migrations run automatically on backend startup

```bash
# Verify migrations exist
ls -la backend/migrations/

# You should see files like:
# 1770125704_add_users_table.up.sql
# 1770125704_add_users_table.down.sql
```

### Issue: API returns 404 for `/api/auth/login`

**Symptoms:**
- Frontend shows "Cannot connect to API"
- Vercel logs show 404s
- Routes are registered in backend

**Solution:**
This typically means the backend API isn't responding. Check:

1. Backend is deployed and running
2. Vercel environment variable `VITE_API_BASE_URL` is correct
3. CORS is properly configured in backend
4. Database connection is working

```bash
# Check backend health
curl https://your-backend-url/health
# Should return 200 OK
```

### Issue: "Database connection failed"

**Symptoms:**
- Backend won't start
- Error: `PostgreSQL connected` doesn't appear in logs

**Solution:**
1. Verify `DATABASE_URL` is correct
2. Check Supabase project is active
3. Ensure firewall allows connections
4. Test connection string locally:

```bash
# Test connection
psql $DATABASE_URL -c "SELECT 1"
```

### Issue: Empty dashboard (no requests showing)

**Symptoms:**
- Frontend loads but shows no data
- Database has tables but queries return nothing

**Solution:**
1. Ensure users are created in `users` table
2. Check leave balances are initialized
3. Verify user has correct role/grade

### Issue: Rules not being applied

**Symptoms:**
- Requests aren't auto-approved
- Rules exist but not executing

**Solution:**
1. Check rules are enabled (not deleted)
2. Verify rule conditions match request
3. Ensure approver is assigned
4. Check approval hierarchy in config

---

## 🔒 Security Features

- ✅ JWT token-based authentication
- ✅ Password hashing (bcrypt)
- ✅ SQL injection prevention
- ✅ CORS protection
- ✅ Rate limiting
- ✅ Audit logging
- ✅ Role-based access control
- ✅ Environment variable isolation

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## Demo

[Smart Rule Based Approval Engine Working.webm](https://github.com/user-attachments/assets/4fd762c3-c7cb-4e06-a47d-8acd4d3da4f0)
