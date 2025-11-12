---
title: "Tutorial 4: BaaS 플랫폼 통합 (Supabase)"
description: "Supabase로 백엔드 개발 속도를 극대화합니다"
duration: "45분"
difficulty: "중급"
tags: [tutorial, baas, supabase, authentication, realtime, storage]
---

# Tutorial 4: BaaS 플랫폼 통합 (Supabase)

이 튜토리얼에서는 Supabase를 활용하여 백엔드 개발 속도를 극대화합니다. 인증, 데이터베이스, 실시간 통신, 파일 저장소를 빠르게 구축하고, Alfred의 BaaS Skills로 모범 사례를 적용합니다.

## 🎯 학습 목표

이 튜토리얼을 완료하면 다음을 할 수 있습니다:

- ✅ Supabase 프로젝트 설정 및 초기화하기
- ✅ Authentication (이메일, 소셜 로그인) 통합하기
- ✅ Realtime subscriptions로 실시간 데이터 동기화하기
- ✅ Row Level Security (RLS)로 데이터 보안 강화하기
- ✅ Storage buckets로 파일 업로드/다운로드 구현하기
- ✅ Edge Functions로 서버리스 API 만들기
- ✅ Alfred의 BaaS Skills로 Best Practices 적용하기

## 📋 사전 요구사항

### 필수 계정

- **Supabase 계정**: [supabase.com](https://supabase.com)에서 무료 가입
- **GitHub 계정**: 소셜 로그인 설정용 (선택)

### 필수 설치

- **Node.js 18+** 또는 **Python 3.11+**
- **MoAI-ADK v0.23.0+**
- **Supabase CLI**: `npm install -g supabase` (선택)

### 선행 지식

- REST API 기본
- JavaScript/TypeScript 기초 또는 Python 기초
- SQL 기본 문법

### 설치 확인

```bash
# 프로젝트 디렉토리
mkdir supabase-chat-app
cd supabase-chat-app

# MoAI-ADK 초기화
moai-adk init

# Supabase CLI 설치 (선택)
npm install -g supabase
```

## 🏗️ 프로젝트 개요: 실시간 채팅 앱

**기능**:
- 사용자 인증 (이메일, GitHub)
- 실시간 메시지
- 파일 첨부 (이미지)
- 온라인 상태 표시

**기술 스택**:
- **Backend**: Supabase (PostgreSQL + Auth + Realtime + Storage)
- **Frontend**: React + TypeScript (선택) 또는 Python FastAPI
- **Deployment**: Vercel (Frontend), Supabase (Backend)

## 🚀 프로젝트 구조

```
supabase-chat-app/
├── .moai/
│   └── specs/
│       └── SPEC-SUPABASE-001.md
├── supabase/
│   ├── migrations/
│   │   └── 20240115_initial_schema.sql
│   ├── functions/
│   │   └── send-notification/
│   └── config.toml
├── src/
│   ├── lib/
│   │   └── supabase.ts         # Supabase 클라이언트
│   ├── components/
│   │   ├── Chat.tsx            # 채팅 컴포넌트
│   │   ├── Auth.tsx            # 인증 컴포넌트
│   │   └── FileUpload.tsx      # 파일 업로드
│   └── hooks/
│       └── useRealtime.ts      # Realtime hook
├── .env.example
├── package.json
└── README.md
```

## 단계별 실습

### Step 1: SPEC 작성

```bash
/alfred:1-plan "Supabase 실시간 채팅 앱"
```

**생성된 SPEC** (`.moai/specs/SPEC-SUPABASE-001.md`):

```markdown
# SPEC-SUPABASE-001: Supabase 실시간 채팅 앱

## 요구사항

Supabase를 활용한 실시간 채팅 애플리케이션

### 기능 요구사항

#### 인증 (Authentication)

- FR-001: 이메일 회원가입/로그인
- FR-002: GitHub OAuth 로그인
- FR-003: 자동 로그인 유지 (세션 관리)
- FR-004: 로그아웃

#### 채팅 (Realtime)

- FR-005: 실시간 메시지 전송/수신
- FR-006: 메시지 목록 조회 (페이지네이션)
- FR-007: 온라인 사용자 표시
- FR-008: 타이핑 인디케이터

#### 파일 첨부 (Storage)

- FR-009: 이미지 업로드 (최대 5MB)
- FR-010: 이미지 미리보기
- FR-011: 파일 다운로드

#### 보안 (RLS)

- SR-001: 로그인한 사용자만 메시지 조회 가능
- SR-002: 본인만 자신의 메시지 수정/삭제 가능
- SR-003: 업로드한 사용자만 파일 삭제 가능

### 데이터 모델

profiles:
- id: uuid (PK, FK to auth.users)
- username: text (unique)
- avatar_url: text
- status: text (online, offline, away)
- updated_at: timestamp

messages:
- id: uuid (PK)
- user_id: uuid (FK to profiles)
- content: text
- file_url: text (nullable)
- created_at: timestamp

### Supabase 기능 활용

- Auth: 이메일 + OAuth (GitHub)
- Realtime: messages 테이블 구독
- Storage: avatars, attachments buckets
- RLS: 모든 테이블에 정책 적용
- Edge Functions: 알림 전송
```

### Step 2: Supabase 프로젝트 생성

1. **Supabase Dashboard 접속**
   - [app.supabase.com](https://app.supabase.com) 로그인
   - "New Project" 클릭

2. **프로젝트 설정**
   ```
   Name: chat-app
   Database Password: 강력한 비밀번호 생성
   Region: Northeast Asia (Seoul) - 가장 가까운 리전 선택
   Pricing Plan: Free (시작용)
   ```

3. **API Keys 확인**
   - Settings → API
   - `Project URL` 복사
   - `anon` key (public) 복사
   - `service_role` key (private) 복사

### Step 3: 환경 설정

**.env.example**:
```env
# Supabase
VITE_SUPABASE_URL=https://your-project.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key

# Service role key (서버 전용, 절대 클라이언트 노출 금지)
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
```

실제 `.env` 파일 생성:
```bash
cp .env.example .env
# Dashboard에서 복사한 값 입력
```

**package.json**:
```json
{
  "name": "supabase-chat-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@supabase/supabase-js": "^2.39.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

설치:
```bash
npm install
```

### Step 4: 데이터베이스 스키마 생성

**supabase/migrations/20240115_initial_schema.sql**:

```sql
-- Supabase 채팅 앱 스키마

-- 1. profiles 테이블 (사용자 프로필)
CREATE TABLE profiles (
    id UUID REFERENCES auth.users(id) PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    avatar_url TEXT,
    status TEXT DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'away')),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. messages 테이블 (채팅 메시지)
CREATE TABLE messages (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES profiles(id) NOT NULL,
    content TEXT NOT NULL,
    file_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. 인덱스 생성 (성능 최적화)
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_profiles_username ON profiles(username);

-- 4. Row Level Security (RLS) 정책

-- profiles: 모든 사용자가 조회 가능, 본인만 수정 가능
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Public profiles are viewable by everyone"
ON profiles FOR SELECT
USING (true);

CREATE POLICY "Users can update own profile"
ON profiles FOR UPDATE
USING (auth.uid() = id);

-- messages: 인증된 사용자만 조회/생성, 본인만 수정/삭제
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Messages are viewable by authenticated users"
ON messages FOR SELECT
TO authenticated
USING (true);

CREATE POLICY "Authenticated users can create messages"
ON messages FOR INSERT
TO authenticated
WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own messages"
ON messages FOR UPDATE
TO authenticated
USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own messages"
ON messages FOR DELETE
TO authenticated
USING (auth.uid() = user_id);

-- 5. 트리거 (자동 프로필 생성)
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.profiles (id, username, avatar_url)
    VALUES (
        NEW.id,
        COALESCE(NEW.raw_user_meta_data->>'username', NEW.email),
        NEW.raw_user_meta_data->>'avatar_url'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
AFTER INSERT ON auth.users
FOR EACH ROW
EXECUTE FUNCTION public.handle_new_user();

-- 6. Realtime 활성화
ALTER PUBLICATION supabase_realtime ADD TABLE messages;
ALTER PUBLICATION supabase_realtime ADD TABLE profiles;
```

**SQL Editor에서 실행**:
1. Supabase Dashboard → SQL Editor
2. 위 SQL 복사/붙여넣기
3. "Run" 클릭

### Step 5: Storage Buckets 생성

**Dashboard에서 설정**:

1. **Storage → Create Bucket**
   - Name: `avatars`
   - Public: ✅ (프로필 사진 공개)

2. **Create Bucket**
   - Name: `attachments`
   - Public: ✅ (첨부 파일 공개)

**RLS 정책 (avatars)**:

```sql
-- Storage Bucket: avatars
-- 모든 사용자가 읽기 가능, 본인만 업로드/수정/삭제

CREATE POLICY "Avatar images are publicly accessible"
ON storage.objects FOR SELECT
USING (bucket_id = 'avatars');

CREATE POLICY "Users can upload own avatar"
ON storage.objects FOR INSERT
WITH CHECK (
    bucket_id = 'avatars' AND
    auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Users can update own avatar"
ON storage.objects FOR UPDATE
USING (
    bucket_id = 'avatars' AND
    auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Users can delete own avatar"
ON storage.objects FOR DELETE
USING (
    bucket_id = 'avatars' AND
    auth.uid()::text = (storage.foldername(name))[1]
);
```

### Step 6: Supabase 클라이언트 초기화

**src/lib/supabase.ts**:

```typescript
/**
 * Supabase 클라이언트 설정
 */
import { createClient } from '@supabase/supabase-js';

const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY;

if (!supabaseUrl || !supabaseAnonKey) {
    throw new Error('Missing Supabase environment variables');
}

export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
    auth: {
        autoRefreshToken: true,
        persistSession: true,
        detectSessionInUrl: true,
    },
    realtime: {
        params: {
            eventsPerSecond: 10,
        },
    },
});

// 타입 정의
export interface Profile {
    id: string;
    username: string;
    avatar_url?: string;
    status: 'online' | 'offline' | 'away';
    updated_at: string;
}

export interface Message {
    id: string;
    user_id: string;
    content: string;
    file_url?: string;
    created_at: string;
    profiles?: Profile;  // JOIN 결과
}
```

### Step 7: 인증 구현

**src/components/Auth.tsx**:

```typescript
/**
 * 인증 컴포넌트
 */
import { useState } from 'react';
import { supabase } from '../lib/supabase';

export function Auth() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(false);
    const [mode, setMode] = useState<'signin' | 'signup'>('signin');

    // 이메일 로그인/회원가입
    const handleEmailAuth = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            if (mode === 'signup') {
                const { data, error } = await supabase.auth.signUp({
                    email,
                    password,
                    options: {
                        data: {
                            username: email.split('@')[0],
                        },
                    },
                });

                if (error) throw error;
                alert('회원가입 성공! 이메일을 확인해주세요.');
            } else {
                const { data, error } = await supabase.auth.signInWithPassword({
                    email,
                    password,
                });

                if (error) throw error;
                console.log('로그인 성공', data);
            }
        } catch (error: any) {
            alert(error.message);
        } finally {
            setLoading(false);
        }
    };

    // GitHub OAuth 로그인
    const handleGithubLogin = async () => {
        const { data, error } = await supabase.auth.signInWithOAuth({
            provider: 'github',
            options: {
                redirectTo: window.location.origin,
            },
        });

        if (error) {
            alert(error.message);
        }
    };

    return (
        <div className="auth-container">
            <h2>{mode === 'signin' ? '로그인' : '회원가입'}</h2>

            <form onSubmit={handleEmailAuth}>
                <input
                    type="email"
                    placeholder="이메일"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                />
                <input
                    type="password"
                    placeholder="비밀번호 (최소 6자)"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    minLength={6}
                />
                <button type="submit" disabled={loading}>
                    {loading ? '처리 중...' : mode === 'signin' ? '로그인' : '회원가입'}
                </button>
            </form>

            <button onClick={() => setMode(mode === 'signin' ? 'signup' : 'signin')}>
                {mode === 'signin' ? '회원가입하기' : '로그인하기'}
            </button>

            <hr />

            <button onClick={handleGithubLogin}>
                GitHub으로 로그인
            </button>
        </div>
    );
}
```

### Step 8: Realtime 채팅 구현

**src/hooks/useRealtime.ts**:

```typescript
/**
 * Realtime 메시지 구독 Hook
 */
import { useEffect, useState } from 'react';
import { supabase, Message } from '../lib/supabase';
import { RealtimeChannel } from '@supabase/supabase-js';

export function useRealtimeMessages() {
    const [messages, setMessages] = useState<Message[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        // 초기 메시지 로드
        loadMessages();

        // Realtime 구독
        const channel: RealtimeChannel = supabase
            .channel('messages')
            .on(
                'postgres_changes',
                {
                    event: 'INSERT',
                    schema: 'public',
                    table: 'messages',
                },
                (payload) => {
                    console.log('새 메시지:', payload.new);
                    // 프로필 정보 포함하여 메시지 추가
                    loadMessageWithProfile(payload.new.id);
                }
            )
            .on(
                'postgres_changes',
                {
                    event: 'DELETE',
                    schema: 'public',
                    table: 'messages',
                },
                (payload) => {
                    console.log('메시지 삭제:', payload.old);
                    setMessages((prev) =>
                        prev.filter((msg) => msg.id !== payload.old.id)
                    );
                }
            )
            .subscribe();

        return () => {
            channel.unsubscribe();
        };
    }, []);

    const loadMessages = async () => {
        setLoading(true);
        const { data, error } = await supabase
            .from('messages')
            .select(`
                *,
                profiles:user_id (
                    id,
                    username,
                    avatar_url
                )
            `)
            .order('created_at', { ascending: true })
            .limit(50);

        if (error) {
            console.error('메시지 로드 실패:', error);
        } else {
            setMessages(data || []);
        }
        setLoading(false);
    };

    const loadMessageWithProfile = async (messageId: string) => {
        const { data } = await supabase
            .from('messages')
            .select(`
                *,
                profiles:user_id (
                    id,
                    username,
                    avatar_url
                )
            `)
            .eq('id', messageId)
            .single();

        if (data) {
            setMessages((prev) => [...prev, data]);
        }
    };

    const sendMessage = async (content: string, fileUrl?: string) => {
        const { data: { user } } = await supabase.auth.getUser();

        if (!user) {
            throw new Error('로그인이 필요합니다');
        }

        const { error } = await supabase.from('messages').insert({
            user_id: user.id,
            content,
            file_url: fileUrl,
        });

        if (error) {
            throw error;
        }
    };

    const deleteMessage = async (messageId: string) => {
        const { error } = await supabase
            .from('messages')
            .delete()
            .eq('id', messageId);

        if (error) {
            throw error;
        }
    };

    return {
        messages,
        loading,
        sendMessage,
        deleteMessage,
        refresh: loadMessages,
    };
}
```

**src/components/Chat.tsx**:

```typescript
/**
 * 채팅 컴포넌트
 */
import { useState, useEffect, useRef } from 'react';
import { useRealtimeMessages } from '../hooks/useRealtime';
import { supabase } from '../lib/supabase';

export function Chat() {
    const [newMessage, setNewMessage] = useState('');
    const [currentUser, setCurrentUser] = useState<any>(null);
    const { messages, loading, sendMessage, deleteMessage } = useRealtimeMessages();
    const messagesEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        // 현재 사용자 정보
        supabase.auth.getUser().then(({ data: { user } }) => {
            setCurrentUser(user);
        });
    }, []);

    useEffect(() => {
        // 자동 스크롤
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages]);

    const handleSendMessage = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!newMessage.trim()) return;

        try {
            await sendMessage(newMessage.trim());
            setNewMessage('');
        } catch (error: any) {
            alert(error.message);
        }
    };

    const handleDeleteMessage = async (messageId: string) => {
        if (confirm('메시지를 삭제하시겠습니까?')) {
            try {
                await deleteMessage(messageId);
            } catch (error: any) {
                alert(error.message);
            }
        }
    };

    if (loading) {
        return <div>메시지 로딩 중...</div>;
    }

    return (
        <div className="chat-container">
            <div className="messages">
                {messages.map((msg) => (
                    <div
                        key={msg.id}
                        className={`message ${
                            msg.user_id === currentUser?.id ? 'own' : 'other'
                        }`}
                    >
                        <div className="message-header">
                            <strong>{msg.profiles?.username}</strong>
                            <span className="timestamp">
                                {new Date(msg.created_at).toLocaleTimeString()}
                            </span>
                        </div>
                        <div className="message-content">{msg.content}</div>
                        {msg.file_url && (
                            <img src={msg.file_url} alt="첨부 파일" />
                        )}
                        {msg.user_id === currentUser?.id && (
                            <button onClick={() => handleDeleteMessage(msg.id)}>
                                삭제
                            </button>
                        )}
                    </div>
                ))}
                <div ref={messagesEndRef} />
            </div>

            <form onSubmit={handleSendMessage} className="message-input">
                <input
                    type="text"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    placeholder="메시지 입력..."
                />
                <button type="submit">전송</button>
            </form>
        </div>
    );
}
```

### Step 9: 파일 업로드 구현

**src/components/FileUpload.tsx**:

```typescript
/**
 * 파일 업로드 컴포넌트
 */
import { useState } from 'react';
import { supabase } from '../lib/supabase';

export function FileUpload({ onUpload }: { onUpload: (url: string) => void }) {
    const [uploading, setUploading] = useState(false);

    const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
        try {
            setUploading(true);

            const file = event.target.files?.[0];
            if (!file) return;

            // 파일 크기 체크 (5MB)
            if (file.size > 5 * 1024 * 1024) {
                alert('파일 크기는 5MB 이하여야 합니다');
                return;
            }

            // 현재 사용자
            const { data: { user } } = await supabase.auth.getUser();
            if (!user) throw new Error('로그인이 필요합니다');

            // 파일명 생성 (중복 방지)
            const fileExt = file.name.split('.').pop();
            const fileName = `${user.id}/${Date.now()}.${fileExt}`;

            // 업로드
            const { data, error } = await supabase.storage
                .from('attachments')
                .upload(fileName, file);

            if (error) throw error;

            // 공개 URL 생성
            const { data: { publicUrl } } = supabase.storage
                .from('attachments')
                .getPublicUrl(fileName);

            onUpload(publicUrl);
        } catch (error: any) {
            alert(error.message);
        } finally {
            setUploading(false);
        }
    };

    return (
        <div>
            <input
                type="file"
                accept="image/*"
                onChange={handleFileUpload}
                disabled={uploading}
            />
            {uploading && <span>업로드 중...</span>}
        </div>
    );
}
```

### Step 10: Edge Function (서버리스 API)

**supabase/functions/send-notification/index.ts**:

```typescript
/**
 * Edge Function: 새 메시지 알림
 */
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts';
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2';

serve(async (req) => {
    try {
        const { message, userId } = await req.json();

        // Supabase 클라이언트 (service_role key 사용)
        const supabase = createClient(
            Deno.env.get('SUPABASE_URL') ?? '',
            Deno.env.get('SUPABASE_SERVICE_ROLE_KEY') ?? ''
        );

        // 사용자 정보 조회
        const { data: profile } = await supabase
            .from('profiles')
            .select('*')
            .eq('id', userId)
            .single();

        // 알림 로직 (예: 이메일, 푸시 알림 등)
        console.log(`새 메시지: ${profile?.username} - ${message}`);

        return new Response(
            JSON.stringify({ success: true }),
            { headers: { 'Content-Type': 'application/json' } }
        );
    } catch (error) {
        return new Response(
            JSON.stringify({ error: error.message }),
            { status: 500, headers: { 'Content-Type': 'application/json' } }
        );
    }
});
```

**배포**:
```bash
supabase functions deploy send-notification
```

## ✅ 테스트 및 검증

### 로컬 실행

```bash
npm run dev
```

브라우저에서 `http://localhost:5173` 열기

### 기능 테스트

1. **회원가입**
   - 이메일 입력 → 이메일 확인 링크 클릭

2. **로그인**
   - 이메일/비밀번호 또는 GitHub OAuth

3. **실시간 채팅**
   - 다른 브라우저/탭에서 동일 사용자 또는 다른 계정 로그인
   - 메시지 전송 → 실시간 동기화 확인

4. **파일 업로드**
   - 이미지 선택 → 업로드 → 메시지에 첨부 확인

5. **RLS 검증**
   - 다른 사용자 메시지 삭제 시도 → 실패 확인

## 🔧 문제 해결

### 문제 1: Realtime 연결 실패

**증상**:
```
RealtimeClient connection failed
```

**해결**:
```typescript
// Supabase Dashboard → Settings → API
// Realtime가 활성화되어 있는지 확인

// 테이블에 Realtime 활성화
ALTER PUBLICATION supabase_realtime ADD TABLE messages;
```

### 문제 2: RLS 정책으로 데이터 조회 안 됨

**증상**: 로그인했는데도 데이터 안 보임

**해결**:
```sql
-- SQL Editor에서 RLS 정책 확인
SELECT * FROM pg_policies WHERE tablename = 'messages';

-- 정책 임시 비활성화 (테스트용)
ALTER TABLE messages DISABLE ROW LEVEL SECURITY;

-- 문제 해결 후 다시 활성화
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
```

### 문제 3: CORS 에러

**증상**:
```
Access-Control-Allow-Origin error
```

**해결**:
```
Supabase Dashboard → Settings → API
→ CORS Settings
→ 로컬 개발 URL 추가: http://localhost:5173
```

## 💡 Best Practices

### 1. RLS 정책 항상 적용

```sql
-- ❌ 나쁜 예: RLS 없음 (보안 취약)
CREATE TABLE posts (...);

-- ✅ 좋은 예: RLS 활성화
CREATE TABLE posts (...);
ALTER TABLE posts ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can view own posts" ON posts FOR SELECT USING (auth.uid() = user_id);
```

### 2. Storage 파일명 중복 방지

```typescript
// ✅ 좋은 예: UUID + timestamp
const fileName = `${user.id}/${crypto.randomUUID()}.${fileExt}`;
```

### 3. Realtime 최적화

```typescript
// 필요한 컬럼만 선택
.select('id, content, created_at')

// 페이지네이션 적용
.order('created_at', { ascending: false })
.limit(50)
```

## 🚀 다음 단계

축하합니다! Supabase로 실시간 앱을 완성했습니다.

### 추가 기능

1. **Presence**: 실시간 온라인 사용자 표시
2. **Broadcast**: 타이핑 인디케이터
3. **Functions**: 이메일 알림, 이미지 리사이징
4. **Webhooks**: 외부 서비스 연동

### 다음 튜토리얼

- **[Tutorial 5: MCP 서버 개발](/ko/tutorials/tutorial-05-mcp-server)**
  - Model Context Protocol 통합

## 📚 참고 자료

- [Supabase 공식 문서](https://supabase.com/docs)
- [Row Level Security](https://supabase.com/docs/guides/auth/row-level-security)
- [Realtime](https://supabase.com/docs/guides/realtime)
- [Storage](https://supabase.com/docs/guides/storage)

---

**Happy Building! 🚀**
