# Realtime Chat Frontend

This is the frontend for the Realtime Chat application built with Next.js 15, TypeScript, and TailwindCSS.

## Features

- ✅ User authentication (login/signup)
- ✅ Channel listing page
- ✅ Real-time chat interface for #general channel
- ✅ WebSocket connection for real-time messaging
- ✅ Message history loading
- ✅ Responsive design with TailwindCSS
- ✅ State management with Zustand
- ✅ Error boundaries and loading states

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript
- **Styling**: TailwindCSS
- **State Management**: Zustand
- **HTTP Client**: Axios
- **Real-time**: WebSocket

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
│   ├── page.tsx           # Home page (redirects)
│   ├── login/page.tsx     # Login/Signup page
│   ├── channels/page.tsx  # Channel listing
│   └── chat/general/page.tsx # Chat interface
├── components/            # Reusable components
│   ├── ErrorBoundary.tsx
│   ├── LoadingSpinner.tsx
│   ├── MessageInput.tsx
│   ├── MessageList.tsx
│   └── ProtectedRoute.tsx
├── hooks/                 # Custom hooks
│   └── useWebSocket.ts
├── lib/                   # API and utilities
│   └── api.ts
├── store/                 # Zustand stores
│   ├── authStore.ts
│   └── chatStore.ts
└── types/                 # TypeScript types
    ├── auth.ts
    └── message.ts
```

## Getting Started

1. Install dependencies:
   ```bash
   npm install
   ```

2. Set up environment variables:
   ```bash
   cp .env.example .env.local
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Environment Variables

- `NEXT_PUBLIC_API_URL`: Backend API URL (default: http://localhost:8080/api)
- `NEXT_PUBLIC_WS_URL`: WebSocket URL (default: ws://localhost:8080)

## Available Scripts

- `npm run dev`: Start development server
- `npm run build`: Build for production
- `npm run start`: Start production server
- `npm run lint`: Run ESLint

## API Integration

The frontend communicates with the Go backend through:

1. **REST API** for authentication and message history
2. **WebSocket** for real-time messaging

## Features Implementation

### Authentication
- JWT-based authentication
- Persistent login state with Zustand
- Automatic token refresh
- Protected routes

### Real-time Chat
- WebSocket connection management
- Message sending and receiving
- Connection status indicators
- Automatic reconnection handling

### UI/UX
- Responsive design for mobile and desktop
- Loading states and error handling
- Message timestamps and user avatars
- Smooth scrolling to new messages
