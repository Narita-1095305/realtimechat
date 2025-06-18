import { User } from './auth';

export interface Message {
  id: number;
  content: string;
  channel: string;
  user_id: number;
  user: User;
  created_at: string;
  updated_at: string;
}

export interface SendMessageRequest {
  content: string;
  channel: string;
}

export interface GetMessagesResponse {
  messages: Message[];
  total: number;
  page: number;
  limit: number;
}