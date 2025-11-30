# Advanced Examples & Production Patterns

## Example 1: Complete User Settings Page

```tsx
import {
 Card,
 CardContent,
 CardDescription,
 CardHeader,
 CardTitle,
} from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import React from "react"

export function UserSettingsPage() {
 const [settings, setSettings] = React.useState({
 emailNotifications: true,
 pushNotifications: false,
 twoFactorAuth: true,
 })

 return (
 <div className="max-w-4xl mx-auto p-6">
 <div className="mb-8">
 <h1 className="text-3xl font-bold mb-2">Account Settings</h1>
 <p className="text-gray-600">Manage your account and preferences</p>
 </div>

 <Tabs defaultValue="general" className="w-full">
 <TabsList className="grid w-full grid-cols-3">
 <TabsTrigger value="general">General</TabsTrigger>
 <TabsTrigger value="notifications">Notifications</TabsTrigger>
 <TabsTrigger value="security">Security</TabsTrigger>
 </TabsList>

 {/* General Settings */}
 <TabsContent value="general">
 <Card>
 <CardHeader>
 <CardTitle>Profile Information</CardTitle>
 <CardDescription>
 Update your profile details
 </CardDescription>
 </CardHeader>
 <CardContent className="space-y-6">
 <div className="grid grid-cols-2 gap-4">
 <div>
 <label className="block text-sm font-medium mb-2">
 First Name
 </label>
 <Input placeholder="John" />
 </div>
 <div>
 <label className="block text-sm font-medium mb-2">
 Last Name
 </label>
 <Input placeholder="Doe" />
 </div>
 </div>

 <div>
 <label className="block text-sm font-medium mb-2">
 Email
 </label>
 <Input type="email" placeholder="john@example.com" />
 </div>

 <div>
 <label className="block text-sm font-medium mb-2">
 Bio
 </label>
 <textarea
 placeholder="Tell us about yourself..."
 className="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
 />
 </div>

 <Button>Save Changes</Button>
 </CardContent>
 </Card>
 </TabsContent>

 {/* Notifications */}
 <TabsContent value="notifications">
 <Card>
 <CardHeader>
 <CardTitle>Notification Preferences</CardTitle>
 <CardDescription>
 Control how you receive notifications
 </CardDescription>
 </CardHeader>
 <CardContent className="space-y-6">
 {[
 {
 key: 'emailNotifications',
 label: 'Email Notifications',
 description: 'Receive email updates'
 },
 {
 key: 'pushNotifications',
 label: 'Push Notifications',
 description: 'Receive push notifications'
 },
 ].map((notif) => (
 <div
 key={notif.key}
 className="flex items-center justify-between py-4 border-b last:border-b-0"
 >
 <div>
 <p className="font-medium">{notif.label}</p>
 <p className="text-sm text-gray-600">
 {notif.description}
 </p>
 </div>
 <Switch
 checked={settings[notif.key]}
 onCheckedChange={(checked) =>
 setSettings({
 ...settings,
 [notif.key]: checked,
 })
 }
 />
 </div>
 ))}

 <Button className="mt-6">Save Preferences</Button>
 </CardContent>
 </Card>
 </TabsContent>

 {/* Security */}
 <TabsContent value="security">
 <Card>
 <CardHeader>
 <CardTitle>Security Settings</CardTitle>
 <CardDescription>
 Manage your account security
 </CardDescription>
 </CardHeader>
 <CardContent className="space-y-6">
 <div className="flex items-center justify-between py-4 border-b">
 <div>
 <p className="font-medium">Two-Factor Authentication</p>
 <p className="text-sm text-gray-600">
 Add extra security to your account
 </p>
 </div>
 <Badge variant={settings.twoFactorAuth ? "default" : "outline"}>
 {settings.twoFactorAuth ? "Enabled" : "Disabled"}
 </Badge>
 </div>

 <div>
 <Button variant="outline">Change Password</Button>
 </div>

 <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
 <h3 className="font-semibold text-red-900 mb-2">
 Danger Zone
 </h3>
 <p className="text-sm text-red-800 mb-4">
 Once you delete your account, there is no going back.
 </p>
 <Button variant="destructive" size="sm">
 Delete Account
 </Button>
 </div>
 </CardContent>
 </Card>
 </TabsContent>
 </Tabs>
 </div>
 )
}
```

## Example 2: Data Table with Sorting & Filtering

```tsx
import {
 Table,
 TableBody,
 TableCell,
 TableHead,
 TableHeader,
 TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
 Select,
 SelectContent,
 SelectItem,
 SelectTrigger,
 SelectValue,
} from "@/components/ui/select"
import React from "react"

interface User {
 id: string
 name: string
 email: string
 status: "active" | "inactive"
 joinDate: string
}

const users: User[] = [
 {
 id: "1",
 name: "Alice Johnson",
 email: "alice@example.com",
 status: "active",
 joinDate: "2025-01-15"
 },
 {
 id: "2",
 name: "Bob Smith",
 email: "bob@example.com",
 status: "active",
 joinDate: "2025-02-20"
 },
 {
 id: "3",
 name: "Charlie Brown",
 email: "charlie@example.com",
 status: "inactive",
 joinDate: "2025-03-10"
 },
]

export function UserTable() {
 const [searchTerm, setSearchTerm] = React.useState("")
 const [statusFilter, setStatusFilter] = React.useState("all")
 const [sortBy, setSortBy] = React.useState("name")

 const filtered = users.filter((user) => {
 const matchesSearch = user.name
 .toLowerCase()
 .includes(searchTerm.toLowerCase())
 const matchesStatus =
 statusFilter === "all" || user.status === statusFilter
 return matchesSearch && matchesStatus
 })

 const sorted = [...filtered].sort((a, b) => {
 switch (sortBy) {
 case "name":
 return a.name.localeCompare(b.name)
 case "date":
 return new Date(b.joinDate).getTime() - new Date(a.joinDate).getTime()
 default:
 return 0
 }
 })

 return (
 <div className="space-y-4">
 {/* Controls */}
 <div className="flex gap-4 flex-wrap">
 <Input
 placeholder="Search users..."
 value={searchTerm}
 onChange={(e) => setSearchTerm(e.target.value)}
 className="flex-1 min-w-[200px]"
 />

 <Select value={statusFilter} onValueChange={setStatusFilter}>
 <SelectTrigger className="w-[150px]">
 <SelectValue />
 </SelectTrigger>
 <SelectContent>
 <SelectItem value="all">All Status</SelectItem>
 <SelectItem value="active">Active</SelectItem>
 <SelectItem value="inactive">Inactive</SelectItem>
 </SelectContent>
 </Select>

 <Select value={sortBy} onValueChange={setSortBy}>
 <SelectTrigger className="w-[150px]">
 <SelectValue />
 </SelectTrigger>
 <SelectContent>
 <SelectItem value="name">Sort by Name</SelectItem>
 <SelectItem value="date">Sort by Date</SelectItem>
 </SelectContent>
 </Select>
 </div>

 {/* Table */}
 <div className="border rounded-lg">
 <Table>
 <TableHeader>
 <TableRow>
 <TableHead>Name</TableHead>
 <TableHead>Email</TableHead>
 <TableHead>Status</TableHead>
 <TableHead>Join Date</TableHead>
 <TableHead>Actions</TableHead>
 </TableRow>
 </TableHeader>
 <TableBody>
 {sorted.map((user) => (
 <TableRow key={user.id}>
 <TableCell className="font-medium">{user.name}</TableCell>
 <TableCell>{user.email}</TableCell>
 <TableCell>
 <span
 className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${
 user.status === "active"
 ? "bg-green-100 text-green-800"
 : "bg-gray-100 text-gray-800"
 }`}
 >
 {user.status}
 </span>
 </TableCell>
 <TableCell>{user.joinDate}</TableCell>
 <TableCell>
 <Button size="sm" variant="ghost">
 Edit
 </Button>
 </TableCell>
 </TableRow>
 ))}
 </TableBody>
 </Table>
 </div>

 {sorted.length === 0 && (
 <div className="text-center py-8 text-gray-500">
 No users found matching your criteria
 </div>
 )}
 </div>
 )
}
```

## Example 3: Responsive Mobile Navigation

```tsx
import { useState } from "react"
import { Button } from "@/components/ui/button"
import {
 Sheet,
 SheetContent,
 SheetDescription,
 SheetHeader,
 SheetTitle,
 SheetTrigger,
} from "@/components/ui/sheet"

export function MobileNav() {
 const [open, setOpen] = useState(false)

 const navItems = [
 { label: "Home", href: "/" },
 { label: "Features", href: "/features" },
 { label: "Pricing", href: "/pricing" },
 { label: "Blog", href: "/blog" },
 { label: "Contact", href: "/contact" },
 ]

 return (
 <nav className="flex items-center justify-between p-4 border-b">
 {/* Logo */}
 <a href="/" className="text-xl font-bold">
 MyApp
 </a>

 {/* Desktop Navigation */}
 <div className="hidden md:flex gap-6">
 {navItems.map((item) => (
 <a
 key={item.label}
 href={item.href}
 className="text-sm hover:text-blue-600 transition-colors"
 >
 {item.label}
 </a>
 ))}
 <Button size="sm">Sign In</Button>
 </div>

 {/* Mobile Menu */}
 <Sheet open={open} onOpenChange={setOpen}>
 <SheetTrigger asChild className="md:hidden">
 <Button variant="ghost" size="icon">
 
 </Button>
 </SheetTrigger>
 <SheetContent side="right">
 <SheetHeader>
 <SheetTitle>Menu</SheetTitle>
 </SheetHeader>
 <div className="space-y-4 mt-6">
 {navItems.map((item) => (
 <a
 key={item.label}
 href={item.href}
 className="block py-2 hover:text-blue-600 transition-colors"
 onClick={() => setOpen(false)}
 >
 {item.label}
 </a>
 ))}
 <Button className="w-full">Sign In</Button>
 </div>
 </SheetContent>
 </Sheet>
 </nav>
 )
}
```

## Example 4: Form with Validation

```tsx
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
 Select,
 SelectContent,
 SelectItem,
 SelectTrigger,
 SelectValue,
} from "@/components/ui/select"
import { Checkbox } from "@/components/ui/checkbox"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import React from "react"

export function SignupForm() {
 const [formData, setFormData] = React.useState({
 email: "",
 password: "",
 confirmPassword: "",
 country: "",
 newsletter: false,
 terms: false,
 })

 const [errors, setErrors] = React.useState({})

 const handleSubmit = (e) => {
 e.preventDefault()
 const newErrors = {}

 if (!formData.email) newErrors.email = "Email is required"
 if (!formData.password) newErrors.password = "Password is required"
 if (formData.password !== formData.confirmPassword) {
 newErrors.confirmPassword = "Passwords do not match"
 }
 if (!formData.country) newErrors.country = "Country is required"
 if (!formData.terms) newErrors.terms = "You must accept the terms"

 setErrors(newErrors)

 if (Object.keys(newErrors).length === 0) {
 console.log("Form submitted:", formData)
 }
 }

 return (
 <form onSubmit={handleSubmit} className="max-w-md mx-auto space-y-4">
 <div>
 <label className="block text-sm font-medium mb-1">Email</label>
 <Input
 type="email"
 placeholder="your@email.com"
 value={formData.email}
 onChange={(e) =>
 setFormData({ ...formData, email: e.target.value })
 }
 />
 {errors.email && (
 <p className="text-red-500 text-sm mt-1">{errors.email}</p>
 )}
 </div>

 <div>
 <label className="block text-sm font-medium mb-1">Password</label>
 <Input
 type="password"
 placeholder="••••••••"
 value={formData.password}
 onChange={(e) =>
 setFormData({ ...formData, password: e.target.value })
 }
 />
 {errors.password && (
 <p className="text-red-500 text-sm mt-1">{errors.password}</p>
 )}
 </div>

 <div>
 <label className="block text-sm font-medium mb-1">
 Confirm Password
 </label>
 <Input
 type="password"
 placeholder="••••••••"
 value={formData.confirmPassword}
 onChange={(e) =>
 setFormData({ ...formData, confirmPassword: e.target.value })
 }
 />
 {errors.confirmPassword && (
 <p className="text-red-500 text-sm mt-1">{errors.confirmPassword}</p>
 )}
 </div>

 <div>
 <label className="block text-sm font-medium mb-1">Country</label>
 <Select value={formData.country} onValueChange={(value) =>
 setFormData({ ...formData, country: value })
 }>
 <SelectTrigger>
 <SelectValue placeholder="Select a country" />
 </SelectTrigger>
 <SelectContent>
 <SelectItem value="usa">United States</SelectItem>
 <SelectItem value="canada">Canada</SelectItem>
 <SelectItem value="uk">United Kingdom</SelectItem>
 </SelectContent>
 </Select>
 {errors.country && (
 <p className="text-red-500 text-sm mt-1">{errors.country}</p>
 )}
 </div>

 <div className="space-y-2">
 <div className="flex items-center space-x-2">
 <Checkbox
 id="newsletter"
 checked={formData.newsletter}
 onCheckedChange={(checked) =>
 setFormData({ ...formData, newsletter: checked })
 }
 />
 <label htmlFor="newsletter" className="text-sm cursor-pointer">
 Subscribe to newsletter
 </label>
 </div>

 <div className="flex items-center space-x-2">
 <Checkbox
 id="terms"
 checked={formData.terms}
 onCheckedChange={(checked) =>
 setFormData({ ...formData, terms: checked })
 }
 />
 <label htmlFor="terms" className="text-sm cursor-pointer">
 I agree to the terms and conditions
 </label>
 </div>

 {errors.terms && (
 <p className="text-red-500 text-sm">{errors.terms}</p>
 )}
 </div>

 <Button type="submit" className="w-full">
 Sign Up
 </Button>
 </form>
 )
}
```
