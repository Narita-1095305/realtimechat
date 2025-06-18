'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/authStore';

export default function ChannelsPage() {
  const { isAuthenticated, user, initializeAuth } = useAuthStore();
  const router = useRouter();
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    const initAuth = async () => {
      console.log('[ChannelsPage] Initializing auth...');
      await initializeAuth();
      setIsInitialized(true);
      console.log('[ChannelsPage] Auth initialized, isAuthenticated:', useAuthStore.getState().isAuthenticated);
    };

    initAuth();
  }, [initializeAuth]);

  useEffect(() => {
    if (isInitialized && !isAuthenticated) {
      console.log('[ChannelsPage] Not authenticated, redirecting to login');
      router.push('/login');
    }
  }, [isAuthenticated, router, isInitialized]);

  if (!isInitialized || (!isAuthenticated && isInitialized)) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <h1 className="text-3xl font-bold text-gray-900">Channels</h1>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-500">Welcome, {user?.username}</span>
              <button
                onClick={() => {
                  useAuthStore.getState().clearAuth();
                  router.push('/login');
                }}
                className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Available Channels
              </h3>
              
              {/* General Channel Card */}
              <div className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 cursor-pointer transition-colors">
                <div 
                  onClick={() => router.push('/chat/general')}
                  className="flex items-center justify-between"
                >
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <div className="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center">
                        <span className="text-indigo-600 font-semibold">#</span>
                      </div>
                    </div>
                    <div className="ml-4">
                      <h4 className="text-lg font-medium text-gray-900">general</h4>
                      <p className="text-sm text-gray-500">
                        Main discussion channel for everyone
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                      Active
                    </span>
                    <svg
                      className="ml-2 w-5 h-5 text-gray-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 5l7 7-7 7"
                      />
                    </svg>
                  </div>
                </div>
              </div>

              <div className="mt-6 text-center text-gray-500">
                <p className="text-sm">
                  More channels coming soon! For now, join the general discussion.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}