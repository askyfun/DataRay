export interface CreateShareRequest {
  chart_id: number;
  password?: string;
  expires_at?: string;
}

export interface ShareResponse {
  id: number;
  token: string;
  chart_id: number;
  password?: string;
  expires_at?: string;
  created_at?: string;
}

export interface ShareListResponse {
  items: ShareResponse[];
  total: number;
}

export interface VerifyPasswordRequest {
  password: string;
}
