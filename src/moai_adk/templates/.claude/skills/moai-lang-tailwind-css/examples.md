# Advanced Examples & Production Patterns

## Example 1: Complete E-Commerce Product Page

```jsx
export function ProductPage() {
  const [quantity, setQuantity] = React.useState(1);
  const [selectedColor, setSelectedColor] = React.useState('black');

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Breadcrumb */}
      <nav className="max-w-6xl mx-auto px-4 py-3 text-sm text-gray-600">
        <a href="/" className="hover:text-gray-900">Home</a>
        <span className="mx-2">/</span>
        <a href="/products" className="hover:text-gray-900">Products</a>
        <span className="mx-2">/</span>
        <span className="text-gray-900">Premium Wireless Headphones</span>
      </nav>

      {/* Product Section */}
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {/* Image Gallery */}
          <div className="lg:col-span-2">
            <img
              src="/product.jpg"
              alt="Premium Wireless Headphones"
              className="w-full rounded-lg shadow-lg mb-4"
            />
            <div className="grid grid-cols-4 gap-2">
              {[1, 2, 3, 4].map((i) => (
                <button
                  key={i}
                  className="border-2 border-gray-200 rounded-lg overflow-hidden hover:border-gray-400 transition-colors"
                >
                  <img
                    src={`/product-${i}.jpg`}
                    alt={`View ${i}`}
                    className="w-full h-20 object-cover"
                  />
                </button>
              ))}
            </div>
          </div>

          {/* Product Info */}
          <div className="flex flex-col">
            <h1 className="text-3xl font-bold text-gray-900 mb-2">
              Premium Wireless Headphones
            </h1>

            {/* Rating */}
            <div className="flex items-center gap-2 mb-4">
              <div className="flex text-yellow-400">
                {[1, 2, 3, 4, 5].map((i) => (
                  <span key={i}>★</span>
                ))}
              </div>
              <span className="text-gray-600 text-sm">(324 reviews)</span>
            </div>

            {/* Price */}
            <div className="mb-6">
              <span className="text-4xl font-bold text-gray-900">$299.99</span>
              <span className="text-lg text-gray-500 line-through ml-2">$399.99</span>
            </div>

            {/* Description */}
            <p className="text-gray-600 mb-6">
              Experience premium audio quality with our noise-cancelling wireless headphones.
              Perfect for work, travel, and everyday use.
            </p>

            {/* Color Selection */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-900 mb-3">
                Color
              </label>
              <div className="flex gap-3">
                {['black', 'silver', 'blue'].map((color) => (
                  <button
                    key={color}
                    onClick={() => setSelectedColor(color)}
                    className={`
                      w-8 h-8 rounded-full border-2 transition-all
                      ${selectedColor === color
                        ? 'border-gray-900 ring-2 ring-offset-2 ring-gray-900'
                        : 'border-gray-300'
                      }
                    `}
                    style={{ backgroundColor: color }}
                    title={color}
                  />
                ))}
              </div>
            </div>

            {/* Quantity & Add to Cart */}
            <div className="flex gap-4 mb-6">
              <div className="flex items-center border border-gray-300 rounded-lg">
                <button
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                  className="px-4 py-2 text-gray-600 hover:text-gray-900"
                >
                  −
                </button>
                <span className="px-4 py-2 border-l border-r border-gray-300">
                  {quantity}
                </span>
                <button
                  onClick={() => setQuantity(quantity + 1)}
                  className="px-4 py-2 text-gray-600 hover:text-gray-900"
                >
                  +
                </button>
              </div>

              <button className="flex-1 bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 font-medium transition-colors">
                Add to Cart
              </button>
            </div>

            {/* Features */}
            <div className="border-t pt-6">
              <h3 className="font-semibold text-gray-900 mb-3">Key Features</h3>
              <ul className="space-y-2 text-gray-600">
                <li className="flex items-center gap-2">
                  <span className="text-green-600">✓</span>
                  Active Noise Cancellation
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-600">✓</span>
                  40-hour Battery Life
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-600">✓</span>
                  Bluetooth 5.0 Connectivity
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
```

## Example 2: Responsive Dashboard Layout

```jsx
export function Dashboard() {
  const [sidebarOpen, setSidebarOpen] = React.useState(true);

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Top Navigation */}
      <nav className="bg-white shadow-sm sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex justify-between items-center">
          <div className="flex items-center gap-4">
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="lg:hidden p-2 rounded-md text-gray-600 hover:bg-gray-100"
            >
              ☰
            </button>
            <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          </div>
          <div className="flex items-center gap-4">
            <input
              type="search"
              placeholder="Search..."
              className="hidden md:block px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button className="p-2 rounded-full bg-gray-100 hover:bg-gray-200">
              🔔
            </button>
            <img src="/avatar.jpg" alt="Profile" className="w-8 h-8 rounded-full" />
          </div>
        </div>
      </nav>

      <div className="flex">
        {/* Sidebar */}
        <aside
          className={`
            ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
            lg:translate-x-0
            fixed lg:static
            left-0 top-16 lg:top-0
            w-64 h-screen
            bg-gray-900 text-white
            transition-transform duration-300
            z-30
          `}
        >
          <nav className="p-6 space-y-2">
            {[
              { label: 'Overview', icon: '📊' },
              { label: 'Analytics', icon: '📈' },
              { label: 'Reports', icon: '📋' },
              { label: 'Settings', icon: '⚙️' },
            ].map((item) => (
              <a
                key={item.label}
                href="#"
                className="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-800 transition-colors"
              >
                <span>{item.icon}</span>
                <span>{item.label}</span>
              </a>
            ))}
          </nav>
        </aside>

        {/* Main Content */}
        <main className="flex-1 p-4 sm:p-8">
          {/* Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            {[
              { label: 'Total Users', value: '12,345', change: '+2.5%' },
              { label: 'Revenue', value: '$45,231', change: '+12.3%' },
              { label: 'Orders', value: '1,234', change: '+5.2%' },
              { label: 'Growth', value: '23.5%', change: '+4.1%' },
            ].map((stat) => (
              <div key={stat.label} className="bg-white rounded-lg shadow p-6">
                <p className="text-gray-600 text-sm">{stat.label}</p>
                <p className="text-3xl font-bold text-gray-900 mt-2">{stat.value}</p>
                <p className="text-green-600 text-sm mt-2">{stat.change}</p>
              </div>
            ))}
          </div>

          {/* Chart Area */}
          <div className="bg-white rounded-lg shadow p-6 mb-8">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Revenue Trend</h2>
            <div className="h-64 bg-gray-50 rounded flex items-end justify-around p-4">
              {[40, 60, 50, 80, 90, 70, 85].map((height, i) => (
                <div
                  key={i}
                  className="bg-blue-500 rounded-t"
                  style={{
                    height: `${(height / 100) * 100}%`,
                    width: '12%'
                  }}
                />
              ))}
            </div>
          </div>

          {/* Table */}
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50 border-b">
                  <tr>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                      Order ID
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                      Customer
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                      Amount
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-gray-900">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {[
                    { id: '#1001', customer: 'John Doe', amount: '$234.50', status: 'Complete' },
                    { id: '#1002', customer: 'Jane Smith', amount: '$567.80', status: 'Pending' },
                    { id: '#1003', customer: 'Bob Wilson', amount: '$123.45', status: 'Complete' },
                  ].map((row) => (
                    <tr key={row.id} className="border-b hover:bg-gray-50">
                      <td className="px-6 py-3 text-gray-900">{row.id}</td>
                      <td className="px-6 py-3 text-gray-900">{row.customer}</td>
                      <td className="px-6 py-3 text-gray-900">{row.amount}</td>
                      <td className="px-6 py-3">
                        <span
                          className={`
                            px-3 py-1 rounded-full text-sm font-medium
                            ${row.status === 'Complete'
                              ? 'bg-green-100 text-green-800'
                              : 'bg-yellow-100 text-yellow-800'
                            }
                          `}
                        >
                          {row.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
```

## Example 3: Animated Card Hover Effect

```jsx
export function AnimatedCard({ title, description, image }) {
  return (
    <div className="group relative overflow-hidden rounded-lg shadow-lg">
      {/* Image */}
      <img
        src={image}
        alt={title}
        className="w-full h-64 object-cover group-hover:scale-110 transition-transform duration-300"
      />

      {/* Overlay */}
      <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/0 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-6">
        <h3 className="text-white text-xl font-bold mb-2">{title}</h3>
        <p className="text-gray-100 text-sm mb-4">{description}</p>
        <button className="self-start px-4 py-2 bg-white text-black font-medium rounded hover:bg-gray-100 transition-colors">
          Learn More
        </button>
      </div>
    </div>
  );
}
```

## Example 4: Responsive Grid System

```jsx
export function GridSystem() {
  return (
    <div className="max-w-6xl mx-auto p-4">
      {/* Auto-fit Grid */}
      <div className="grid auto-rows-[minmax(200px,1fr)] gap-4 mb-8">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div
            key={i}
            className="bg-gradient-to-br from-blue-400 to-blue-600 rounded-lg flex items-center justify-center text-white text-2xl font-bold shadow-lg"
          >
            {i}
          </div>
        ))}
      </div>

      {/* Responsive Breakpoints */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
          <div key={i} className="aspect-square bg-gray-200 rounded-lg flex items-center justify-center">
            {i}
          </div>
        ))}
      </div>
    </div>
  );
}
```

## Example 5: Utility Classes for Layouts

```jsx
// Flexbox Utilities
<div className="flex flex-col gap-4">
  <div className="flex items-center justify-between">
    <span>Title</span>
    <span>Value</span>
  </div>
</div>

// Grid Utilities
<div className="grid grid-cols-12 gap-4">
  <div className="col-span-8">Main Content</div>
  <div className="col-span-4">Sidebar</div>
</div>

// Spacing Utilities
<div className="px-4 py-6 md:px-8 md:py-12">
  Responsive padding
</div>

// Typography Utilities
<h1 className="text-3xl md:text-5xl font-bold text-gray-900">
  Responsive Heading
</h1>

// Color & Background Utilities
<div className="bg-gradient-to-r from-purple-500 to-pink-500 text-white p-8 rounded-lg">
  Gradient Background
</div>
```
