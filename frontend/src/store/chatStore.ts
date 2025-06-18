import { create } from 'zustand';
import { Message } from '@/types/message';
import { messageApi } from '@/lib/api';

interface ChatState {
  messages: Message[];
  isConnected: boolean;
  currentChannel: string | null;
  ws: WebSocket | null;
  
  // Actions
  connect: (channel: string) => void;
  disconnect: () => void;
  sendMessage: (channel: string, content: string) => Promise<void>;
  addMessage: (message: Message) => void;
  setMessages: (messages: Message[]) => void;
  fetchMessages: (channel: string) => Promise<void>;
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: [],
  isConnected: false,
  currentChannel: null,
  ws: null,

  connect: (channel: string) => {
    const { ws: existingWs, disconnect } = get();
    
    // 既存の接続を完全にクリーンアップ
    if (existingWs) {
      console.log('🔄 Cleaning up existing WebSocket connection');
      disconnect();
      // 少し待ってから新しい接続を作成
      setTimeout(() => {
        get().connect(channel);
      }, 100);
      return;
    }

    const token = localStorage.getItem('auth_token');
    console.log('🔍 Debug: Token check:', {
      tokenExists: !!token,
      tokenLength: token?.length || 0,
      tokenPreview: token ? `${token.substring(0, 20)}...` : 'null'
    });
    
    if (!token) {
      console.error('❌ No auth token found in localStorage');
      return;
    }

    // トークンの形式を確認（JWT形式かどうか）
    const tokenParts = token.split('.');
    if (tokenParts.length !== 3) {
      console.error('❌ Invalid JWT token format - expected 3 parts, got', tokenParts.length);
      return;
    }

    // JWTペイロードをデコードして有効期限をチェック
    try {
      const payload = JSON.parse(atob(tokenParts[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      
      console.log('🔍 Token payload check:', {
        exp: payload.exp,
        currentTime,
        isExpired: payload.exp && payload.exp < currentTime,
        timeUntilExpiry: payload.exp ? payload.exp - currentTime : 'unknown',
        fullPayload: payload,
        userId: payload.user_id,
        username: payload.username,
        email: payload.email,
        issuer: payload.iss,
        subject: payload.sub
      });
      
      if (payload.exp && payload.exp < currentTime) {
        console.error('❌ Token has expired');
        // 期限切れトークンを削除
        localStorage.removeItem('auth_token');
        return;
      }
    } catch (error) {
      console.error('❌ Failed to decode token payload:', error);
      return;
    }

    const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/ws/chat?token=${token}&channel=${channel}`;
    console.log('🔗 Attempting WebSocket connection to:', wsUrl.replace(token, 'TOKEN_HIDDEN'));
    
    const newWs = new WebSocket(wsUrl);

    newWs.onopen = () => {
      console.log('✅ WebSocket connected successfully');
      set({ isConnected: true, currentChannel: channel, ws: newWs });
    };

    newWs.onmessage = (event) => {
      console.log('📨 Raw WebSocket message received:', {
        data: event.data,
        type: typeof event.data,
        length: event.data.length
      });
      
      // 空のデータや無効なデータの早期チェック
      if (!event.data || event.data.trim() === '' || event.data === '{}') {
        console.warn('⚠️ Empty or invalid WebSocket message ignored:', event.data);
        return;
      }
      
      try {
        const parsedData = JSON.parse(event.data);
        console.log('📨 Parsed WebSocket message:', {
          parsed: parsedData,
          type: typeof parsedData,
          keys: Object.keys(parsedData || {}),
          messageType: parsedData.type,
          hasData: !!parsedData.data
        });
        
        // 空のオブジェクトの場合はスキップ
        if (!parsedData || Object.keys(parsedData).length === 0) {
          console.warn('⚠️ Empty parsed object ignored:', parsedData);
          return;
        }
        
        // メッセージタイプによる処理分岐
        switch (parsedData.type) {
          case 'chat_message':
            if (parsedData.data) {
              console.log('✅ Processing chat message:', parsedData.data);
              
              // WebSocketから受信したメッセージを追加
              // （API経由での保存をやめたので、自分のメッセージもWebSocketから受信する）
              get().addMessage(parsedData.data);
            } else {
              console.warn('⚠️ Chat message without data:', parsedData);
            }
            break;
            
          case 'system':
            console.log('ℹ️ System message received:', parsedData);
            break;
            
          case 'user_joined':
          case 'user_left':
            console.log('👥 User status message:', parsedData);
            break;
            
          case 'pong':
            console.log('🏓 Pong received:', parsedData);
            break;
            
          case 'users_list':
            console.log('👥 Users list received:', parsedData);
            break;
            
          default:
            console.log('⚠️ Unknown message type received:', parsedData);
        }
      } catch (error) {
        console.error('❌ Failed to parse WebSocket message:', error);
        console.error('❌ Raw data that failed to parse:', event.data);
      }
    };

    newWs.onclose = (event) => {
      console.log('🔌 WebSocket disconnected:', {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean
      });
      set({ isConnected: false, ws: null });
    };

    newWs.onerror = (error) => {
      console.error('❌ WebSocket error occurred');
      console.error('Error details:', {
        type: error.type || 'unknown',
        readyState: newWs.readyState,
        url: wsUrl.replace(token, 'TOKEN_HIDDEN'),
        timestamp: new Date().toISOString(),
        errorEvent: error.constructor.name
      });
      
      // WebSocketの状態も確認
      const readyStateNames = ['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED'];
      console.error('WebSocket state:', readyStateNames[newWs.readyState] || newWs.readyState);
      
      set({ isConnected: false });
    };
  },

  disconnect: () => {
    const { ws } = get();
    if (ws) {
      ws.close();
    }
    set({ isConnected: false, currentChannel: null, ws: null });
  },

  sendMessage: async (channel: string, content: string) => {
    const { ws } = get();
    
    console.log('🔍 sendMessage called:', {
      channel,
      content,
      wsExists: !!ws,
      wsReadyState: ws?.readyState,
      wsOpen: ws?.readyState === WebSocket.OPEN
    });
    
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      const error = 'WebSocket is not connected';
      console.error('❌ WebSocket not ready:', {
        ws: !!ws,
        readyState: ws?.readyState,
        expected: WebSocket.OPEN
      });
      throw new Error(error);
    }

    try {
      // WebSocket経由でメッセージを送信（バックエンドで保存・配信処理）
      const messageData = {
        type: 'chat_message',
        channel,
        content,
      };

      console.log('📤 Sending WebSocket message:', messageData);
      ws.send(JSON.stringify(messageData));
      console.log('✅ WebSocket message sent successfully');
      
    } catch (error) {
      console.error('❌ Failed to send message:', error);
      throw error;
    }
  },

  addMessage: (message: Message) => {
    // デバッグ: メッセージの構造確認
    console.log('📝 Adding message:', {
      hasId: !!message.id,
      id: message.id,
      content: message.content,
      structure: Object.keys(message)
    });
    
    if (!message.id) {
      console.error('❌ Message received without ID:', message);
      return; // IDのないメッセージは追加しない
    }
    
    set((state) => {
      // 重複チェック: 同じIDのメッセージが既に存在するかチェック
      const existingMessage = state.messages.find(m => m.id === message.id);
      if (existingMessage) {
        console.warn('⚠️ Duplicate message ignored:', {
          id: message.id,
          content: message.content
        });
        return state; // 重複メッセージは追加しない
      }
      
      return {
        messages: [...state.messages, message].sort((a, b) => 
          new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
        ),
      };
    });
  },

  setMessages: (messages: Message[]) => {
    set({ messages: messages.sort((a, b) => 
      new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    ) });
  },

  fetchMessages: async (channel: string) => {
    try {
      const messages = await messageApi.getMessages(channel);
      get().setMessages(messages);
    } catch (error) {
      console.error('Failed to fetch messages:', error);
      throw error;
    }
  },
}));