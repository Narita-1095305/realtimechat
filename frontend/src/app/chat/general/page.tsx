'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/authStore';
import { useChatStore } from '@/store/chatStore';
import { MessageInput } from '@/components/MessageInput';
import { MessageList } from '@/components/MessageList';

export default function GeneralChatPage() {
  const { isAuthenticated, user } = useAuthStore();
  const { 
    messages, 
    isConnected, 
    connect, 
    disconnect, 
    sendMessage,
    fetchMessages 
  } = useChatStore();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // 認証状態を初期化
    const { initializeAuth } = useAuthStore.getState();
    initializeAuth();

    // 少し待ってから認証チェック（Zustand persistの復元を待つ）
    const checkAuth = () => {
      const { isAuthenticated } = useAuthStore.getState();
      console.log('🔍 Auth check result:', { isAuthenticated });
      
      if (!isAuthenticated) {
        console.log('❌ Not authenticated, redirecting to login');
        router.push('/login');
        return;
      }

      const initializeChat = async () => {
        try {
          // Fetch message history first
          await fetchMessages('general');
          // Then connect to WebSocket
          connect('general');
          setIsLoading(false);
        } catch (error) {
          console.error('Failed to initialize chat:', error);
          setIsLoading(false);
        }
      };

      initializeChat();
    };

    // 初回は即座にチェック、その後100ms後にも再チェック（persist復元のため）
    checkAuth();
    const timeoutId = setTimeout(checkAuth, 100);

    // Cleanup on unmount
    return () => {
      clearTimeout(timeoutId);
      disconnect();
    };
  }, [router, connect, disconnect, fetchMessages]);

  const handleSendMessage = async (content: string) => {
    console.log('🚀 handleSendMessage called with content:', content);
    console.log('🚀 Current user:', user);
    console.log('🚀 IsConnected:', isConnected);
    
    if (!content.trim() || !user) {
      console.warn('⚠️ Message sending cancelled:', {
        contentEmpty: !content.trim(),
        userMissing: !user
      });
      return;
    }
    
    try {
      console.log('📤 Attempting to send message via WebSocket...');
      await sendMessage('general', content);
      console.log('✅ Message sent successfully');
    } catch (error) {
      console.error('❌ Failed to send message:', error);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Redirecting to login...</div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading chat...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-4">
              <button
                onClick={() => router.push('/channels')}
                className="text-gray-500 hover:text-gray-700"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </button>
              <div className="flex items-center">
                <span className="text-xl font-semibold text-gray-900"># general</span>
                <div className={`ml-3 w-3 h-3 rounded-full ${isConnected ? 'bg-green-400' : 'bg-red-400'}`}></div>
                <span className="ml-2 text-sm text-gray-500">
                  {isConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-500">Welcome, {user?.username}</span>
              <button
                onClick={() => {
                  useAuthStore.getState().clearAuth();
                  router.push('/login');
                }}
                className="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-sm"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Chat Area */}
      <div className="flex-1 flex flex-col max-w-4xl mx-auto w-full">
        {/* Messages */}
        <div className="flex-1 overflow-hidden">
          <MessageList messages={messages} currentUser={user} />
        </div>

        {/* Message Input */}
        <div className="border-t bg-white p-4">
          <MessageInput 
            onSendMessage={handleSendMessage}
            disabled={!isConnected}
            placeholder={isConnected ? "Type a message..." : "Connecting..."}
          />
        </div>
      </div>
    </div>
  );
}