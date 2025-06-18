'use client';

import { useEffect, useRef } from 'react';
import { Message } from '@/types/message';
import { User } from '@/types/auth';

interface MessageListProps {
  messages: Message[];
  currentUser: User | null;
}

export function MessageList({ messages, currentUser }: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const formatDate = (timestamp: string) => {
    const date = new Date(timestamp);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    if (date.toDateString() === today.toDateString()) {
      return 'Today';
    } else if (date.toDateString() === yesterday.toDateString()) {
      return 'Yesterday';
    } else {
      return date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: date.getFullYear() !== today.getFullYear() ? 'numeric' : undefined,
      });
    }
  };

  const shouldShowDateSeparator = (currentMessage: Message, previousMessage?: Message) => {
    if (!previousMessage) return true;
    
    const currentDate = new Date(currentMessage.created_at).toDateString();
    const previousDate = new Date(previousMessage.created_at).toDateString();
    
    return currentDate !== previousDate;
  };

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center text-gray-500">
          <div className="text-6xl mb-4">💬</div>
          <h3 className="text-lg font-medium mb-2">Welcome to #general!</h3>
          <p className="text-sm">This is the beginning of the #general channel.</p>
          <p className="text-sm">Send a message to get the conversation started.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {messages.map((message, index) => {
        // デバッグ: メッセージIDの状態確認
        if (message.id === undefined || message.id === null) {
          console.warn('⚠️ Message without ID found at index:', index, message);
        }
        
        const previousMessage = index > 0 ? messages[index - 1] : undefined;
        const showDateSeparator = shouldShowDateSeparator(message, previousMessage);
        const isOwnMessage = currentUser?.id === message.user_id;

        return (
          <div key={message.id ? `message-${message.id}` : `message-fallback-${index}-${message.created_at}`}>
            {showDateSeparator && (
              <div className="flex items-center justify-center my-6">
                <div className="bg-gray-100 text-gray-600 text-xs px-3 py-1 rounded-full">
                  {formatDate(message.created_at)}
                </div>
              </div>
            )}
            
            <div className={`flex ${isOwnMessage ? 'justify-end' : 'justify-start'}`}>
              <div className={`max-w-xs lg:max-w-md ${isOwnMessage ? 'order-2' : 'order-1'}`}>
                <div
                  className={`px-4 py-2 rounded-lg ${
                    isOwnMessage
                      ? 'bg-indigo-600 text-white'
                      : 'bg-white border border-gray-200 text-gray-900'
                  }`}
                >
                  {!isOwnMessage && (
                    <div className="text-xs font-medium text-gray-600 mb-1">
                      {message.user.username}
                    </div>
                  )}
                  <div className="text-sm whitespace-pre-wrap break-words">
                    {message.content}
                  </div>
                  <div
                    className={`text-xs mt-1 ${
                      isOwnMessage ? 'text-indigo-200' : 'text-gray-500'
                    }`}
                  >
                    {formatTime(message.created_at)}
                  </div>
                </div>
              </div>
              
              {!isOwnMessage && (
                <div className="order-1 mr-3 flex-shrink-0">
                  <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                    <span className="text-xs font-medium text-gray-600">
                      {message.user.username.charAt(0).toUpperCase()}
                    </span>
                  </div>
                </div>
              )}
            </div>
          </div>
        );
      })}
      <div ref={messagesEndRef} />
    </div>
  );
}