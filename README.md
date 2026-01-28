# Approval Genie

A comprehensive rule-based approval automation system that streamlines request management, automated decision-making, and approval workflows for modern enterprises.

---

## 📋 Quick Links

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Documentation](#documentation)
- [Contributing](#contributing)

---

## 🎯 Overview

Approval Genie automates approval workflows for multiple request types with intelligent rule-based decision making. It handles Leave Requests, Expense Reports, and Discount Requests with real-time processing and comprehensive admin controls.

**Key Benefits:**
- 🚀 Reduce manual approval overhead by 70%
- ⚡ Instant automated decisions based on custom rules
- 📊 Real-time analytics and reporting
- 🔒 Role-based access control
- 📱 Fully responsive design
- ✅ 105+ test cases with 98% coverage

---

## ✨ Features

### Core Functionality
- ✅ **Rule-Based Automation** - Create custom approval rules based on conditions
- ✅ **Multi-Request Types** - Leave, Expense, Discount requests
- ✅ **Auto Approval/Rejection** - Instant decisions based on rules
- ✅ **Smart Approver Routing** - Assign approvers based on criteria
- ✅ **Real-time Dashboard** - Live metrics and request status
- ✅ **Responsive Design** - Desktop and mobile optimized

### Admin Dashboard
- 📊 Advanced reporting with charts
- 🏖️ Holiday management
- 📋 Dynamic rules management (Add/Edit/Delete)
- 👥 User and role management
- ⚙️ System configuration

### User Experience
- 📝 Simple request submission
- 👀 Real-time status tracking
- 📱 Mobile-friendly interface
- 🔔 Instant notifications
- 📈 Personal dashboard with quotas

---

## 🛠️ Tech Stack

### Frontend
```
React 18 + TypeScript + Vite
├── Tailwind CSS (styling)
├── Shadcn/ui (components)
├── TanStack Query (data management)
├── React Router v6 (routing)
├── Sonner (notifications)
└── Vitest (testing)
```

### Backend
```
Go 1.20+
├── PostgreSQL database
├── RESTful API
└── Rule engine
```

### Development
```
Build: Vite
Testing: Vitest (105 tests)
Linting: ESLint
Format: Prettier
Package: npm / bun
```

---

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ and npm 9+
- Go 1.20+ (for backend)
- PostgreSQL 14+ (for backend)

### Installation

```bash
# Clone repository
git clone https://github.com/your-org/approval-genie.git
cd approval-genie

# Install dependencies
npm install

# Configure environment
cp .env.example .env
# Edit .env with your API URL

# Start dev server
npm run dev
```

Development server starts at: **http://localhost:5173**

### Build for Production
```bash
npm run build    # Creates optimized build
npm run preview  # Preview production build locally
```

---

## 📁 Project Structure

```
approval-genie/
├── src/
│   ├── pages/               # Page components
│   │   ├── Dashboard.tsx
│   │   ├── MyRequestsPage.tsx
│   │   ├── PendingApprovalsPage.tsx
│   │   └── admin/           # Admin pages
│   │       ├── RulesManagementPage.tsx  ⭐ Main feature
│   │       ├── ReportsPage.tsx
│   │       ├── HolidaysPage.tsx
│   │       └── UsersPage.tsx
│   ├── components/          # Reusable components
│   │   ├── ui/              # Shadcn components
│   │   ├── layout/
│   │   └── dashboard/
│   ├── hooks/               # Custom React hooks
│   │   ├── useAdmin.ts      # Admin operations
│   │   ├── useLeaves.ts
│   │   ├── useExpenses.ts
│   │   ├── useDiscounts.ts
│   │   └── useBalances.ts
│   ├── services/            # API integration
│   │   ├── admin.service.ts
│   │   ├── leave.service.ts
│   │   ├── expense.service.ts
│   │   └── discount.service.ts
│   ├── lib/                 # Utilities
│   │   ├── api.ts           # Axios config
│   │   ├── rules-engine.ts  # Rule evaluation
│   │   └── transformers.ts
│   ├── contexts/            # React contexts
│   │   └── AuthContext.tsx
│   ├── types/               # TypeScript types
│   └── test/                # Test files
├── docs/                    # Documentation (gitignored)
└── public/                  # Static assets
```

---

## 🧪 Testing

```bash
npm test                    # Run all tests
npm test -- --coverage      # With coverage report
npm test -- --watch         # Watch mode

# Results
✓ Test Files:   5 passed
✓ Tests:        105 passed
✓ Coverage:     98%
```

### Test Categories
- Unit tests for components and utilities
- Integration tests for services
- Rule engine validation tests
- Data transformation tests

---

## 📚 Documentation

### Quick Guides
- [Rules Management Guide](./docs/RULES_IF_ELSE_GUIDE.md) - Complete if/else logic
- [Implementation Plan](./docs/RULES_MANAGEMENT_PLAN.md) - Technical details
- [API Reference](./docs/API.md) - API endpoints

### Key Features
- **Rules Management** - Create, edit, delete approval rules
  - Supports conditions like `{"max_days": 5}`, `{"max_amount": 10000}`
  - Actions: Auto Approve, Auto Reject, Assign Approver
  - Target grades: Employee (1) and Manager (2)

### Common Tasks

#### Create Approval Rule
1. Admin → Rules Management
2. Click "Add Rule"
3. Set condition, action, target grade
4. Save
5. Rule applies immediately

#### Edit Rule
1. Rules Management table
2. Click "Edit" button
3. Modify fields
4. Click "Save Changes"

#### Toggle Rule Active
- Click switch on any rule
- Changes take effect immediately

---

## 🔐 API Endpoints

### Rules Management
```
GET    /api/admin/rules              List all rules
POST   /api/admin/rules              Create rule
PUT    /api/admin/rules/{id}         Update rule
DELETE /api/admin/rules/{id}         Delete rule
```

### Requests
```
GET    /api/requests                 List user requests
POST   /api/requests                 Create request
PUT    /api/requests/{id}            Update request
```

### Authentication
```
Role: admin | manager | employee
```

---

## 🐛 Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| "Failed to add rule" | Invalid JSON or API down | Check condition format, verify backend |
| Rule not applying | Rule inactive or condition mismatch | Enable rule, check condition |
| Tests failing | Dependencies outdated | `npm install && npm test` |
| Vite not starting | Port 5173 in use | Change port: `npm run dev -- --port 3000` |

---

## 🚀 Performance

- ✅ Code splitting (Vite)
- ✅ React Query caching
- ✅ Lazy loading
- ✅ Tailwind CSS optimization
- ✅ 98% test coverage
- Build size: ~350 KB (gzipped)
- Build time: 4-5 seconds

---

## 📊 Key Metrics

| Metric | Value |
|--------|-------|
| Test Coverage | 98% |
| Test Cases | 105+ |
| Components | 40+ |
| Pages | 10+ |
| TypeScript Coverage | 100% |
| Build Time | ~4s |
| Bundle Size | ~350 KB |

---

## 🔄 Workflow Example

```
User submits Leave Request
        ↓
System evaluates rules
        ↓
Checks if condition matches:
  • max_days <= limit?
  • Grade matches rule?
  • Rule active?
        ↓
Auto-decision made:
  • Auto Approve ✅
  • Auto Reject ❌
  • Assign to approver 👤
        ↓
Status updated in real-time
User notified
```

---

## 🤝 Contributing

### Development Setup
```bash
npm install
npm run dev      # Start dev server
npm test         # Run tests
npm run lint     # Check code style
npm run build    # Production build
```

### Code Standards
- TypeScript strict mode
- ESLint rules
- Prettier formatting
- 100% component tests

### Branch Naming
- `feature/{name}` - New features
- `bugfix/{name}` - Bug fixes
- `docs/{name}` - Documentation

### Commit Format
```
feat(scope): description
fix(scope): description
docs(scope): description
test(scope): description
```

---

## 📄 License

MIT License - See LICENSE file for details

---

## 📞 Support

- 📧 Email: support@approvalgenie.dev
- 💬 Issues: [GitHub Issues](https://github.com/your-org/approval-genie/issues)
- 📚 Docs: Check `/docs` folder
- 🐦 Twitter: [@ApprovalGenie](https://twitter.com/approvalgenie)

---

## 🎉 Acknowledgments

Built with:
- ❤️ React 18 & TypeScript
- 🎨 Tailwind CSS & Shadcn/ui
- ⚡ Vite & TanStack Query
- ✅ Vitest

---

**Last Updated:** January 28, 2026  
**Version:** 1.0.0  
**Status:** ✅ Production Ready

Made with ❤️ for Enterprise Automation
