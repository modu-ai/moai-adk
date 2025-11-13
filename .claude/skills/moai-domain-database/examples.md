# Database Architecture Examples

## Real-World Implementation Examples

### 1. E-commerce Database Architecture

#### Complete Schema Design

```sql
-- E-commerce database schema for high-volume transactions
-- PostgreSQL 17 optimized structure

-- Categories hierarchy
CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    parent_id BIGINT REFERENCES categories(id),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Products table with comprehensive indexing
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    short_description TEXT,
    price DECIMAL(10,2) NOT NULL,
    compare_price DECIMAL(10,2),
    cost_price DECIMAL(10,2),
    weight DECIMAL(8,3),
    dimensions JSONB,
    category_id BIGINT REFERENCES categories(id),
    brand_id BIGINT REFERENCES brands(id),
    status VARCHAR(20) DEFAULT 'active',
    track_inventory BOOLEAN DEFAULT true,
    meta_title VARCHAR(255),
    meta_description TEXT,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Product variants for size, color, etc.
CREATE TABLE product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255),
    price DECIMAL(10,2) NOT NULL,
    compare_price DECIMAL(10,2),
    cost_price DECIMAL(10,2),
    weight DECIMAL(8,3),
    barcode VARCHAR(100),
    inventory_count INTEGER DEFAULT 0,
    inventory_policy VARCHAR(20) DEFAULT 'deny',
    requires_shipping BOOLEAN DEFAULT true,
    taxable BOOLEAN DEFAULT true,
    position INTEGER DEFAULT 0,
    option1 VARCHAR(100),
    option2 VARCHAR(100),
    option3 VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Customers with comprehensive data
CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    birth_date DATE,
    gender VARCHAR(10),
    accepts_marketing BOOLEAN DEFAULT false,
    tax_exempt BOOLEAN DEFAULT false,
    default_address_id BIGINT,
    total_spent DECIMAL(12,2) DEFAULT 0,
    orders_count INTEGER DEFAULT 0,
    last_order_id BIGINT,
    state VARCHAR(20) DEFAULT 'enabled',
    currency VARCHAR(3) DEFAULT 'USD',
    notes TEXT,
    tags TEXT[],
    metafields JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Addresses with proper normalization
CREATE TABLE addresses (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    company VARCHAR(255),
    address1 VARCHAR(255) NOT NULL,
    address2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    province VARCHAR(100) NOT NULL,
    country VARCHAR(2) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    phone VARCHAR(50),
    default BOOLEAN DEFAULT false,
    address_type VARCHAR(20) DEFAULT 'shipping',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Orders with comprehensive tracking
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_number VARCHAR(100) UNIQUE NOT NULL,
    customer_id BIGINT REFERENCES customers(id),
    financial_status VARCHAR(20) NOT NULL,
    fulfillment_status VARCHAR(20) DEFAULT 'unfulfilled',
    currency VARCHAR(3) NOT NULL,
    total_price DECIMAL(12,2) NOT NULL,
    subtotal_price DECIMAL(12,2) NOT NULL,
    total_tax DECIMAL(10,2) NOT NULL,
    total_shipping DECIMAL(10,2) NOT NULL,
    total_discount DECIMAL(10,2) NOT NULL,
    total_weight DECIMAL(10,3),
    processed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancel_reason VARCHAR(100),
    cancellation_reason TEXT,
    refund_status VARCHAR(20),
    notes TEXT,
    tags TEXT[],
    shipping_address JSONB,
    billing_address JSONB,
    metafields JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Order line items
CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id),
    variant_id BIGINT REFERENCES product_variants(id),
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(500) NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    total_discount DECIMAL(10,2) NOT NULL,
    properties JSONB,
    product_exists BOOLEAN DEFAULT true,
    taxable BOOLEAN DEFAULT true,
    requires_shipping BOOLEAN DEFAULT true,
    gift_card BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Inventory transactions
CREATE TABLE inventory_transactions (
    id BIGSERIAL PRIMARY KEY,
    variant_id BIGINT NOT NULL REFERENCES product_variants(id),
    order_id BIGINT REFERENCES orders(id),
    customer_id BIGINT REFERENCES customers(id),
    location_id BIGINT REFERENCES locations(id),
    quantity_change INTEGER NOT NULL,
    transaction_type VARCHAR(20) NOT NULL, -- 'sale', 'return', 'adjustment', 'transfer'
    reason TEXT,
    cost_price DECIMAL(10,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Performance Optimization for E-commerce

```sql
-- Comprehensive indexing strategy for e-commerce queries

-- Product search and filtering indexes
CREATE INDEX idx_products_category_status ON products(category_id, status) WHERE status = 'active';
CREATE INDEX idx_products_price_range ON products(price) WHERE status = 'active';
CREATE INDEX idx_products_brand ON products(brand_id) WHERE status = 'active';
CREATE INDEX idx_products_created_at ON products(created_at DESC) WHERE status = 'active';
CREATE INDEX idx_products_tags ON products USING GIN(tags) WHERE status = 'active';

-- Customer and order indexes
CREATE INDEX idx_customers_email ON customers(email) WHERE state = 'enabled';
CREATE INDEX idx_customers_orders_count ON customers(orders_count DESC);
CREATE INDEX idx_customers_total_spent ON customers(total_spent DESC);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_customer_status ON orders(customer_id, financial_status);
CREATE INDEX idx_orders_status_date ON orders(financial_status, created_at DESC);

-- Inventory tracking indexes
CREATE INDEX idx_inventory_variant_date ON inventory_transactions(variant_id, created_at DESC);
CREATE INDEX idx_inventory_order_variant ON inventory_transactions(order_id, variant_id);

-- Address and location indexes
CREATE INDEX idx_addresses_customer ON addresses(customer_id);
CREATE INDEX idx_addresses_customer_default ON addresses(customer_id, default) WHERE default = true;

-- Partial indexes for common queries
CREATE INDEX idx_active_products_name_search ON products USING GIN(to_tsvector('english', name)) WHERE status = 'active';
CREATE INDEX idx_products_popular ON products(id) WHERE status = 'active' AND total_sales > 100;
CREATE INDEX idx_variants_low_stock ON product_variants(id, inventory_count) WHERE track_inventory = true AND inventory_count < 10;
```

#### Advanced E-commerce Queries

```sql
-- Customer segmentation and analytics
WITH customer_segments AS (
  SELECT 
    id,
    email,
    total_spent,
    orders_count,
    created_at,
    CASE 
      WHEN total_spent > 1000 AND orders_count > 10 THEN 'VIP'
      WHEN total_spent > 500 AND orders_count > 5 THEN 'Premium'
      WHEN total_spent > 100 AND orders_count > 2 THEN 'Regular'
      ELSE 'New'
    END as segment,
    AVG(order_total) as avg_order_value,
    MAX(created_at) as last_order_date
  FROM (
    SELECT 
      c.id,
      c.email,
      c.total_spent,
      c.orders_count,
      c.created_at,
      o.total_price as order_total,
      o.created_at as order_date
    FROM customers c
    LEFT JOIN orders o ON c.id = o.customer_id
    WHERE c.state = 'enabled'
  ) customer_data
  GROUP BY id, email, total_spent, orders_count, created_at
)
SELECT 
  segment,
  COUNT(*) as customer_count,
  SUM(total_spent) as total_revenue,
  AVG(avg_order_value) as avg_order_value,
  AVG(DAYS(NOW() - last_order_date)) as days_since_last_order
FROM customer_segments
GROUP BY segment
ORDER BY total_revenue DESC;

-- Product performance analysis
WITH product_performance AS (
  SELECT 
    p.id,
    p.name,
    p.category_id,
    c.name as category_name,
    COUNT(oi.id) as total_orders,
    SUM(oi.quantity) as total_quantity_sold,
    SUM(oi.quantity * oi.price) as total_revenue,
    AVG(oi.price) as avg_selling_price,
    p.cost_price,
    SUM(oi.quantity * (oi.price - p.cost_price)) as total_profit,
    RANK() OVER (PARTITION BY p.category_id ORDER BY SUM(oi.quantity * oi.price) DESC) as rank_in_category,
    RANK() OVER (ORDER BY SUM(oi.quantity * oi.price) DESC) as overall_rank
  FROM products p
  JOIN categories c ON p.category_id = c.id
  JOIN order_items oi ON p.id = oi.product_id
  JOIN orders o ON oi.order_id = o.id
  WHERE o.financial_status = 'paid'
    AND o.created_at >= NOW() - INTERVAL '90 days'
  GROUP BY p.id, p.name, p.category_id, c.name, p.cost_price
)
SELECT 
  category_name,
  name,
  total_orders,
  total_quantity_sold,
  total_revenue,
  avg_selling_price,
  total_profit,
  rank_in_category,
  overall_rank,
  ROUND((total_profit / total_revenue) * 100, 2) as profit_margin_percentage
FROM product_performance
WHERE total_revenue > 0
ORDER BY total_revenue DESC
LIMIT 100;

-- Inventory optimization recommendations
WITH inventory_analysis AS (
  SELECT 
    pv.id,
    pv.sku,
    p.name as product_name,
    pv.inventory_count,
    COALESCE(SUM(oi.quantity), 0) as sold_last_30_days,
    COALESCE(SUM(oi.quantity), 0) as sold_last_90_days,
    p.cost_price,
    AVG(oi.price) as avg_selling_price,
    pv.updated_at as last_inventory_update
  FROM product_variants pv
  JOIN products p ON pv.product_id = p.id
  LEFT JOIN order_items oi ON pv.id = oi.variant_id
  LEFT JOIN orders o ON oi.order_id = o.id 
    AND o.financial_status = 'paid'
    AND o.created_at >= NOW() - INTERVAL '90 days'
  WHERE pv.track_inventory = true
  GROUP BY pv.id, pv.sku, p.name, pv.inventory_count, p.cost_price, pv.updated_at
)
SELECT 
  sku,
  product_name,
  inventory_count,
  sold_last_30_days,
  sold_last_90_days,
  CASE 
    WHEN inventory_count = 0 AND sold_last_30_days > 0 THEN 'OUT OF STOCK - URGENT'
    WHEN inventory_count < sold_last_30_days THEN 'LOW STOCK - REORDER SOON'
    WHEN inventory_count > sold_last_90_days * 3 THEN 'OVERSTOCKED - CONSIDER PROMOTION'
    WHEN sold_last_90_days = 0 AND inventory_count > 0 THEN 'SLOW MOVING - CONSIDER DISCOUNT'
    ELSE 'STOCK LEVEL OK'
  END as inventory_status,
  ROUND((sold_last_30_days::DECIMAL / 30), 2) as daily_sales_rate,
  CASE 
    WHEN sold_last_30_days > 0 THEN 
      FLOOR(inventory_count::DECIMAL / (sold_last_30_days::DECIMAL / 30))
    ELSE 999
  END as days_of_stock_remaining
FROM inventory_analysis
ORDER BY 
  CASE 
    WHEN inventory_count = 0 AND sold_last_30_days > 0 THEN 1
    WHEN inventory_count < sold_last_30_days THEN 2
    ELSE 3
  END,
  sold_last_30_days DESC;
```

### 2. Social Media Platform Database Design

#### MongoDB Schema for Social Media

```javascript
// Users collection with comprehensive profile data
db.users.createIndex({"username": 1}, {unique: true});
db.users.createIndex({"email": 1}, {unique: true});
db.users.createIndex({"created_at": -1});
db.users.createIndex({"followers_count": -1});
db.users.createIndex({"location.city": "2dsphere"});

db.users.insertMany([
  {
    _id: ObjectId(),
    username: "john_doe",
    email: "john@example.com",
    password_hash: "$2b$12$...",
    profile: {
      first_name: "John",
      last_name: "Doe",
      bio: "Software engineer and tech enthusiast",
      avatar_url: "https://cdn.example.com/avatars/john.jpg",
      cover_url: "https://cdn.example.com/covers/john.jpg",
      location: {
        city: "San Francisco",
        country: "USA",
        coordinates: [-122.4194, 37.7749]
      },
      website: "https://johndoe.dev",
      birth_date: ISODate("1990-05-15"),
      gender: "male"
    },
    stats: {
      followers_count: 1250,
      following_count: 450,
      posts_count: 325,
      likes_count: 5420
    },
    preferences: {
      private_account: false,
      show_birth_date: false,
      allow_mentions: true,
      allow_direct_messages: true,
      language: "en",
      timezone: "America/Los_Angeles"
    },
    verification: {
      is_verified: true,
      verification_type: "blue_check",
      verification_date: ISODate("2023-01-15")
    },
    status: "active",
    last_active: ISODate("2025-01-13T10:30:00Z"),
    created_at: ISODate("2020-03-10T15:20:00Z"),
    updated_at: ISODate("2025-01-13T09:15:00Z")
  }
]);

// Posts collection with rich content support
db.posts.createIndex({"user_id": 1, "created_at": -1});
db.posts.createIndex({"hashtags": 1});
db.posts.createIndex({"mentions": 1});
db.posts.createIndex({"location.coordinates": "2dsphere"});
db.posts.createIndex({"content": "text", "hashtags": "text"}, {
  weights: {
    "content": 10,
    "hashtags": 5
  },
  name: "post_search_index"
});

// Posts with various content types
db.posts.insertOne({
  _id: ObjectId(),
  user_id: ObjectId("..."),
  content: {
    text: "Just launched my new open source project! Check it out and let me know what you think 🚀 #opensource #coding #github",
    media: [
      {
        type: "image",
        url: "https://cdn.example.com/posts/image1.jpg",
        thumbnail_url: "https://cdn.example.com/posts/thumb1.jpg",
        width: 1920,
        height: 1080,
        size: 245760,
        format: "jpeg"
      }
    ],
    poll: {
      question: "What programming language should I use for my next project?",
      options: [
        { text: "Python", votes: 45 },
        { text: "JavaScript", votes: 32 },
        { text: "Rust", votes: 28 },
        { text: "Go", votes: 15 }
      ],
      total_votes: 120,
      expires_at: ISODate("2025-01-20T10:00:00Z"),
      multiple_choice: false
    }
  },
  hashtags: ["opensource", "coding", "github"],
  mentions: ["@github", "@mozilla"],
  location: {
    name: "San Francisco, CA",
    coordinates: [-122.4194, 37.7749]
  },
  stats: {
    likes_count: 85,
    comments_count: 23,
    shares_count: 12,
    views_count: 1520
  },
  visibility: "public", // public, private, followers_only
  reply_to: null, // ObjectId if this is a reply
  quote_post: null, // ObjectId if this quotes another post
  content_warning: null,
  is_pinned: false,
  status: "published", // draft, published, deleted
  created_at: ISODate("2025-01-13T09:00:00Z"),
  updated_at: ISODate("2025-01-13T09:00:00Z")
});

// Comments/replies collection
db.comments.createIndex({"post_id": 1, "created_at": -1});
db.comments.createIndex({"user_id": 1, "created_at": -1});
db.comments.createIndex({"parent_id": 1, "created_at": -1});

db.comments.insertOne({
  _id: ObjectId(),
  post_id: ObjectId("..."),
  user_id: ObjectId("..."),
  parent_id: null, // For nested comments
  content: {
    text: "This looks amazing! Congrats on the launch 🎉"
  },
  stats: {
    likes_count: 12,
    replies_count: 3
  },
  reply_to: ObjectId("..."), // If replying to specific comment
  status: "published",
  created_at: ISODate("2025-01-13T09:15:00Z"),
  updated_at: ISODate("2025-01-13T09:15:00Z")
});

// Follows/relationships collection
db.follows.createIndex({"follower_id": 1, "following_id": 1}, {unique: true});
db.follows.createIndex({"following_id": 1});
db.follows.createIndex({"follower_id": 1});

db.follows.insertOne({
  _id: ObjectId(),
  follower_id: ObjectId("user_1"),
  following_id: ObjectId("user_2"),
  created_at: ISODate("2025-01-10T14:30:00Z")
});

// Likes collection (for posts, comments, etc.)
db.likes.createIndex({"user_id": 1, "target_id": 1, "target_type": 1}, {unique: true});
db.likes.createIndex({"target_id": 1, "target_type": 1});

db.likes.insertOne({
  _id: ObjectId(),
  user_id: ObjectId("user_1"),
  target_id: ObjectId("post_123"),
  target_type: "post", // post, comment
  created_at: ISODate("2025-01-13T09:05:00Z")
});

// Notifications collection
db.notifications.createIndex({"user_id": 1, "created_at": -1});
db.notifications.createIndex({"read": 1, "created_at": -1});

db.notifications.insertOne({
  _id: ObjectId(),
  user_id: ObjectId("user_1"),
  type: "like", // like, comment, follow, mention, share
  actor_id: ObjectId("user_2"),
  target_id: ObjectId("post_123"),
  target_type: "post",
  message: "Jane Doe liked your post",
  data: {
    actor_username: "jane_doe",
    actor_avatar: "https://cdn.example.com/avatars/jane.jpg",
    post_preview: "Just launched my new open source project..."
  },
  read: false,
  created_at: ISODate("2025-01-13T09:05:00Z")
});
```

#### Advanced Social Media Queries

```javascript
// Generate personalized feed with content ranking
function generatePersonalizedFeed(userId, limit = 20) {
  const pipeline = [
    // Get users that the current user follows
    {
      $lookup: {
        from: "follows",
        localField: "_id",
        foreignField: "following_id",
        as: "following"
      }
    },
    {
      $match: {
        _id: userId
      }
    },
    {
      $project: {
        following_ids: "$following.follower_id"
      }
    },
    {
      $lookup: {
        from: "posts",
        let: { following_ids: "$following_ids" },
        pipeline: [
          {
            $match: {
              $expr: {
                $in: ["$user_id", "$$following_ids"]
              },
              status: "published",
              visibility: { $in: ["public", "followers_only"] }
            }
          },
          {
            $addFields: {
              // Calculate engagement score
              engagement_score: {
                $add: [
                  { $multiply: ["$stats.likes_count", 3] },
                  { $multiply: ["$stats.comments_count", 5] },
                  { $multiply: ["$stats.shares_count", 8] },
                  "$stats.views_count"
                ]
              },
              // Time decay factor (recent posts get higher score)
              time_decay: {
                $divide: [
                  1,
                  { $add: [
                    { $divide: [
                      { $subtract: [new Date(), "$created_at"] },
                      1000 * 60 * 60 * 24 // Convert milliseconds to days
                    ]},
                    1
                  ]}
                ]
              },
              // Personalization based on user's interests
              personalization_score: {
                $cond: {
                  if: { $gt: [{ $size: { $setIntersection: ["$hashtags", ["opensource", "coding", "github"]] } }, 0] },
                  then: 2.5,
                  else: 1
                }
              }
            }
          },
          {
            $addFields: {
              final_score: {
                $multiply: [
                  "$engagement_score",
                  "$time_decay",
                  "$personalization_score"
                ]
              }
            }
          },
          {
            $sort: { final_score: -1 }
          },
          {
            $limit: limit
          },
          {
            $lookup: {
              from: "users",
              localField: "user_id",
              foreignField: "_id",
              as: "user"
            }
          },
          {
            $lookup: {
              from: "likes",
              let: { post_id: "$_id", user_id: userId },
              pipeline: [
                {
                  $match: {
                    $expr: {
                      $and: [
                        { $eq: ["$target_id", "$$post_id"] },
                        { $eq: ["$target_type", "post"] },
                        { $eq: ["$user_id", "$$user_id"] }
                      ]
                    }
                  }
                }
              ],
              as: "user_like"
            }
          },
          {
            $project: {
              content: 1,
              hashtags: 1,
              mentions: 1,
              location: 1,
              stats: 1,
              created_at: 1,
              user: { $arrayElemAt: ["$user", 0] },
              user_liked: { $gt: [{ $size: "$user_like" }, 0] }
            }
          }
        ],
        as: "feed_posts"
      }
    }
  ];
  
  return db.users.aggregate(pipeline).toArray();
}

// Trending topics analysis
function getTrendingHashtags(timeframeHours = 24) {
  const since = new Date(Date.now() - timeframeHours * 60 * 60 * 1000);
  
  const pipeline = [
    {
      $match: {
        created_at: { $gte: since },
        status: "published"
      }
    },
    {
      $unwind: "$hashtags"
    },
    {
      $group: {
        _id: "$hashtags",
        post_count: { $sum: 1 },
        total_likes: { $sum: "$stats.likes_count" },
        total_comments: { $sum: "$stats.comments_count" },
        total_shares: { $sum: "$stats.shares_count" },
        unique_users: { $addToSet: "$user_id" }
      }
    },
    {
      $addFields: {
        user_count: { $size: "$unique_users" },
        engagement_score: {
          $add: [
            { $multiply: ["$total_likes", 3] },
            { $multiply: ["$total_comments", 5] },
            { $multiply: ["$total_shares", 8] }
          ]
        }
      }
    },
    {
      $addFields: {
        trend_score: {
          $multiply: [
            { $ln: { $add: ["$user_count", 1] } }, // Diversity factor
            { $ln: { $add: ["$post_count", 1] } }, // Volume factor
            { $pow: [{ $divide: ["$engagement_score", "$post_count"] }, 0.5] } // Engagement density
          ]
        }
      }
    },
    {
      $sort: { trend_score: -1 }
    },
    {
      $limit: 20
    },
    {
      $project: {
        hashtag: "$_id",
        post_count: 1,
        user_count: 1,
        avg_engagement: { $divide: ["$engagement_score", "$post_count"] },
        trend_score: 1
      }
    }
  ];
  
  return db.posts.aggregate(pipeline).toArray();
}

// User influence and network analysis
function analyzeUserInfluence(userId) {
  const pipeline = [
    {
      $match: { _id: userId }
    },
    {
      $lookup: {
        from: "follows",
        let: { user_id: "$_id" },
        pipeline: [
          {
            $match: {
              $expr: { $eq: ["$following_id", "$$user_id"] }
            }
          },
          {
            $lookup: {
              from: "users",
              localField: "follower_id",
              foreignField: "_id",
              as: "follower_info"
            }
          }
        ],
        as: "followers"
      }
    },
    {
      $lookup: {
        from: "follows",
        let: { user_id: "$_id" },
        pipeline: [
          {
            $match: {
              $expr: { $eq: ["$follower_id", "$$user_id"] }
            }
          }
        ],
        as: "following"
      }
    },
    {
      $lookup: {
        from: "posts",
        let: { user_id: "$_id" },
        pipeline: [
          {
            $match: {
              $expr: { $eq: ["$user_id", "$$user_id"] },
              status: "published",
              created_at: { $gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000) } // Last 30 days
            }
          },
          {
            $group: {
              _id: null,
              total_posts: { $sum: 1 },
              total_likes: { $sum: "$stats.likes_count" },
              total_comments: { $sum: "$stats.comments_count" },
              total_shares: { $sum: "$stats.shares_count" },
              avg_likes_per_post: { $avg: "$stats.likes_count" },
              avg_comments_per_post: { $avg: "$stats.comments_count" }
            }
          }
        ],
        as: "post_stats"
      }
    },
    {
      $addFields: {
        follower_count: { $size: "$followers" },
        following_count: { $size: "$following" },
        post_stats: { $arrayElemAt: ["$post_stats", 0] }
      }
    },
    {
      $addFields: {
        // Calculate influence score
        reach_score: { $ln: { $add: ["$follower_count", 1] } },
        engagement_score: {
          $cond: {
            if: { $gt: ["$post_stats.total_posts", 0] },
            then: {
              $divide: [
                { $add: ["$post_stats.total_likes", "$post_stats.total_comments"] },
                "$post_stats.total_posts"
              ]
            },
            else: 0
          }
        },
        network_quality: {
          $avg: {
            $map: {
              input: "$followers",
              as: "follower",
              in: { $ln: { $add: [{ $size: "$$follower.follower_info.followers" }, 1] } }
            }
          }
        }
      }
    },
    {
      $addFields: {
        influence_score: {
          $multiply: [
            "$reach_score",
            { $add: ["$engagement_score", 1] },
            { $sqrt: "$network_quality" }
          ]
        }
      }
    },
    {
      $project: {
        username: 1,
        profile: 1,
        follower_count: 1,
        following_count: 1,
        post_stats: 1,
        reach_score: 1,
        engagement_score: 1,
        network_quality: 1,
        influence_score: 1
      }
    }
  ];
  
  return db.users.aggregate(pipeline).toArray();
}
```

### 3. Real-time Analytics Database Setup

#### Redis Configuration for Real-time Analytics

```python
# real_time_analytics.py - Comprehensive analytics with Redis
import redis
import json
import time
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import asyncio

class RealTimeAnalytics:
    def __init__(self, redis_host='localhost', redis_port=6379, db=0):
        self.redis = redis.Redis(
            host=redis_host, 
            port=redis_port, 
            db=db,
            decode_responses=True,
            socket_keepalive=True,
            socket_keepalive_options={}
        )
        
        # Initialize data structures
        self._setup_analytics_structures()
    
    def _setup_analytics_structures(self):
        """Setup Redis data structures for analytics"""
        # Time-series data streams
        self.streams = {
            'page_views': 'analytics:page_views',
            'user_events': 'analytics:user_events', 
            'conversions': 'analytics:conversions',
            'errors': 'analytics:errors'
        }
        
        # Aggregated counters
        self.counters = {
            'daily_active_users': 'analytics:dau',
            'page_views_today': 'analytics:page_views:today',
            'conversions_today': 'analytics:conversions:today',
            'error_rate': 'analytics:error_rate'
        }
        
        # Sorted sets for leaderboards
        self.leaderboards = {
            'top_pages': 'analytics:top_pages',
            'top_users': 'analytics:top_users',
            'top_referrers': 'analytics:top_referrers'
        }
        
        # HyperLogLog for unique counts
        self.hll = {
            'unique_visitors': 'analytics:unique_visitors',
            'unique_conversions': 'analytics:unique_conversions'
        }
    
    async def track_page_view(self, user_id: str, page_url: str, referrer: str = None, 
                            user_agent: str = None, session_id: str = None):
        """Track a page view event"""
        timestamp = int(time.time() * 1000)  # milliseconds
        event_data = {
            'user_id': user_id,
            'page_url': page_url,
            'referrer': referrer or '',
            'user_agent': user_agent or '',
            'session_id': session_id or '',
            'timestamp': timestamp,
            'date': datetime.now().strftime('%Y-%m-%d'),
            'hour': datetime.now().hour
        }
        
        # Add to stream
        await self.redis.xadd(
            self.streams['page_views'], 
            event_data,
            maxlen=1000000  # Keep last 1M events
        )
        
        # Update counters
        await asyncio.gather(
            # Daily counters
            self.redis.incr(f"{self.counters['page_views_today']}:{event_data['date']}"),
            self.redis.incr(f"{self.counters['daily_active_users']}:{event_data['date']}"),
            
            # Hourly counters
            self.redis.incr(f"analytics:page_views:hourly:{event_data['date']}:{event_data['hour']}"),
            
            # Page leaderboard
            self.redis.zincrby(self.leaderboards['top_pages'], 1, page_url),
            
            # Referrer leaderboard
            self.redis.zincrby(self.leaderboards['top_referrers'], 1, referrer or 'direct'),
            
            # Unique visitors
            self.redis.pfadd(self.hll['unique_visitors'], user_id),
            
            # User activity tracking
            self.redis.zadd(f"analytics:user_activity:{user_id}", {timestamp: page_url}),
            
            # Session tracking
            self.redis.expire(f"analytics:user_activity:{user_id}", 3600 * 24)  # 24 hours
        )
        
        # Set session expiration
        if session_id:
            await self.redis.expire(f"analytics:session:{session_id}", 1800)  # 30 minutes
    
    async def track_conversion(self, user_id: str, conversion_type: str, 
                             value: float = 0, currency: str = 'USD', 
                             properties: Dict = None):
        """Track a conversion event"""
        timestamp = int(time.time() * 1000)
        event_data = {
            'user_id': user_id,
            'conversion_type': conversion_type,
            'value': str(value),
            'currency': currency,
            'timestamp': timestamp,
            'date': datetime.now().strftime('%Y-%m-%d'),
            'hour': datetime.now().hour,
            'properties': json.dumps(properties or {})
        }
        
        # Add to stream
        await self.redis.xadd(
            self.streams['conversions'],
            event_data,
            maxlen=500000  # Keep last 500K conversions
        )
        
        # Update counters and leaderboards
        await asyncio.gather(
            # Conversion counters
            self.redis.incr(f"{self.counters['conversions_today']}:{event_data['date']}"),
            self.redis.incrby(f"analytics:conversions:value:{event_data['date']}", int(value * 100)),  # Store as cents
            self.redis.incr(f"analytics:conversions:type:{conversion_type}:{event_data['date']}"),
            
            # Unique conversions
            self.redis.pfadd(self.hll['unique_conversions'], f"{user_id}:{conversion_type}"),
            
            # User leaderboard (by conversion value)
            self.redis.zincrby(self.leaderboards['top_users'], value, user_id)
        )
    
    async def track_error(self, user_id: str, error_type: str, error_message: str, 
                         context: Dict = None):
        """Track an error event"""
        timestamp = int(time.time() * 1000)
        event_data = {
            'user_id': user_id,
            'error_type': error_type,
            'error_message': error_message[:500],  # Truncate long messages
            'timestamp': timestamp,
            'date': datetime.now().strftime('%Y-%m-%d'),
            'hour': datetime.now().hour,
            'context': json.dumps(context or {})
        }
        
        # Add to stream
        await self.redis.xadd(
            self.streams['errors'],
            event_data,
            maxlen=100000  # Keep last 100K errors
        )
        
        # Update error counters
        await asyncio.gather(
            self.redis.incr(f"analytics:errors:today:{event_data['date']}"),
            self.redis.incr(f"analytics:errors:type:{error_type}:{event_data['date']}"),
            self.redis.incr(f"analytics:error_rate:{event_data['date']}")
        )
    
    async def get_real_time_metrics(self) -> Dict:
        """Get current real-time metrics"""
        today = datetime.now().strftime('%Y-%m-%d')
        current_hour = datetime.now().hour
        
        # Get basic counters
        basic_metrics = await asyncio.gather(
            self.redis.get(f"{self.counters['page_views_today']}:{today}"),
            self.redis.get(f"{self.counters['conversions_today']}:{today}"),
            self.redis.get(f"analytics:errors:today:{today}"),
            self.redis.get(f"analytics:page_views:hourly:{today}:{current_hour}")
        )
        
        # Get unique counts using HyperLogLog
        unique_metrics = await asyncio.gather(
            self.redis.pfcount(self.hll['unique_visitors']),
            self.redis.pfcount(self.hll['unique_conversions'])
        )
        
        # Get top items from leaderboards
        top_pages = await self.redis.zrevrange(self.leaderboards['top_pages'], 0, 9, withscores=True)
        top_referrers = await self.redis.zrevrange(self.leaderboards['top_referrers'], 0, 4, withscores=True)
        
        return {
            'timestamp': datetime.now().isoformat(),
            'date': today,
            'hour': current_hour,
            'page_views': {
                'today': int(basic_metrics[0] or 0),
                'current_hour': int(basic_metrics[3] or 0)
            },
            'conversions': {
                'today': int(basic_metrics[1] or 0)
            },
            'errors': {
                'today': int(basic_metrics[2] or 0)
            },
            'unique_metrics': {
                'visitors': unique_metrics[0],
                'conversions': unique_metrics[1]
            },
            'top_pages': [{'url': url, 'views': int(views)} for url, views in top_pages],
            'top_referrers': [{'source': source, 'visits': int(visits)} for source, visits in top_referrers]
        }
    
    async def get_user_activity(self, user_id: str, hours: int = 24) -> List[Dict]:
        """Get user activity for the last N hours"""
        since_timestamp = int((time.time() - hours * 3600) * 1000)
        
        # Get user activity from sorted set
        user_activity = await self.redis.zrangebyscore(
            f"analytics:user_activity:{user_id}",
            since_timestamp,
            '+inf',
            withscores=True
        )
        
        return [
            {
                'timestamp': int(timestamp),
                'datetime': datetime.fromtimestamp(timestamp / 1000).isoformat(),
                'page_url': page_url
            }
            for page_url, timestamp in user_activity
        ]
    
    async def get_time_series_data(self, metric: str, hours: int = 24) -> List[Dict]:
        """Get time series data for a specific metric"""
        end_time = datetime.now()
        start_time = end_time - timedelta(hours=hours)
        
        # Read from stream
        stream_name = self.streams.get(metric, f"analytics:{metric}")
        events = await self.redis.xread(
            {stream_name: '$'},
            count=10000,
            block=100
        )
        
        # Aggregate by hour
        hourly_data = {}
        for stream, messages in events:
            for message_id, fields in messages:
                hour_key = f"{fields.get('date', 'unknown')}:{fields.get('hour', 'unknown')}"
                if hour_key not in hourly_data:
                    hourly_data[hour_key] = {
                        'date': fields.get('date', 'unknown'),
                        'hour': int(fields.get('hour', 0)),
                        'count': 0,
                        'unique_users': set()
                    }
                
                hourly_data[hour_key]['count'] += 1
                if 'user_id' in fields:
                    hourly_data[hour_key]['unique_users'].add(fields['user_id'])
        
        # Convert sets to counts and sort
        result = []
        for data in hourly_data.values():
            data['unique_users_count'] = len(data['unique_users'])
            del data['unique_users']
            result.append(data)
        
        return sorted(result, key=lambda x: (x['date'], x['hour']))
    
    async def get_conversion_funnel(self, funnel_steps: List[str], days: int = 7) -> Dict:
        """Analyze conversion funnel performance"""
        end_date = datetime.now()
        start_date = end_date - timedelta(days=days)
        
        funnel_data = {}
        
        for step in funnel_steps:
            step_count = 0
            step_users = set()
            
            # Read conversion events for this step
            events = await self.redis.xread(
                {self.streams['conversions']: 0},
                count=50000
            )
            
            for stream, messages in events:
                for message_id, fields in messages:
                    event_date = datetime.fromtimestamp(int(fields['timestamp']) / 1000)
                    
                    if (start_date <= event_date <= end_date and 
                        fields.get('conversion_type') == step):
                        step_count += 1
                        if 'user_id' in fields:
                            step_users.add(fields['user_id'])
            
            funnel_data[step] = {
                'total_conversions': step_count,
                'unique_users': len(step_users),
                'avg_value': 0  # Would need to calculate from value field
            }
        
        # Calculate conversion rates between steps
        previous_count = None
        for step, data in funnel_data.items():
            if previous_count is not None and previous_count > 0:
                data['conversion_rate'] = (data['total_conversions'] / previous_count) * 100
            else:
                data['conversion_rate'] = 100.0  # First step
            previous_count = data['total_conversions']
        
        return {
            'funnel_steps': funnel_steps,
            'data': funnel_data,
            'analysis_period': f"{days} days",
            'generated_at': datetime.now().isoformat()
        }
```

These comprehensive examples demonstrate real-world database architecture patterns across different domains, providing practical implementations that you can adapt for your specific needs.
