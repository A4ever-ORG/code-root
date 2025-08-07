# Usage Guide - Telegram Store Hub

This guide explains how to use the Telegram Store Hub system for both users and administrators.

## Table of Contents
- [Getting Started](#getting-started)
- [User Guide](#user-guide)
- [Store Management](#store-management)
- [Admin Guide](#admin-guide)
- [Sub-Bot Features](#sub-bot-features)
- [Troubleshooting](#troubleshooting)

## Getting Started

### First Time Setup
1. **Start the Bot**: Send `/start` to your mother bot
2. **Join Channel**: If force join is enabled, join the required channel
3. **Register Store**: Use "ثبت فروشگاه جدید" to create your first store

### Bot Commands
- `/start` - Start interaction with the bot
- `/panel` - Access your store management panel
- `/admin` - Admin panel (admin users only)

## User Guide

### Registering a New Store

#### Step 1: Store Information
1. Click "🏪 ثبت فروشگاه جدید"
2. Enter your store name (minimum 3 characters)
3. Provide a store description

#### Step 2: Plan Selection
Choose from three available plans:

**🆓 Free Plan**
- 10 products maximum
- Fixed button layout
- 5% commission on sales
- Basic support

**💎 Pro Plan (50,000 Toman)**
- 200 products maximum
- Advanced reporting
- Welcome messages
- Custom advertisements
- 5% commission on sales
- Priority support

**👑 VIP Plan (150,000 Toman)**
- Unlimited products
- Dedicated payment gateway
- 0% commission
- Premium advertisements
- Custom button layouts
- 24/7 support

#### Step 3: Payment (Pro/VIP Plans)
1. View payment instructions
2. Transfer money to the provided card number
3. Send payment receipt photo
4. Wait for admin approval

### Managing Your Stores

#### Accessing Store Panel
1. Send `/panel` or click "📊 فروشگاه‌های من"
2. Select a store to manage
3. Access the management panel

#### Store Panel Features
- **➕ افزودن محصول**: Add new products
- **📦 لیست محصولات**: View and manage products
- **🛒 سفارش‌ها**: View customer orders
- **📈 گزارش فروش**: Sales reports
- **🔄 تمدید پلن**: Renew subscription
- **⚙️ تنظیمات فروشگاه**: Store settings

## Store Management

### Adding Products

#### Product Information Required
1. **Product Name**: Clear, descriptive name
2. **Description**: Detailed product description
3. **Price**: Price in Toman (numbers only)
4. **Image**: Product photo (optional)

#### Step-by-Step Process
1. Click "➕ افزودن محصول"
2. Enter product name
3. Provide product description
4. Set the price (e.g., 25000)
5. Upload product image or skip
6. Confirm product creation

### Managing Products

#### Product List
- View all your products
- Check product status (active/inactive)
- Edit or delete products

#### Editing Products
1. Click "📦 لیست محصولات"
2. Select a product
3. Choose edit option:
   - ✏️ Edit name
   - 💰 Edit price
   - 📝 Edit description
   - 🖼 Change image
   - ✅/❌ Toggle availability

#### Deleting Products
1. Select product from list
2. Click "🗑 حذف"
3. Confirm deletion
⚠️ **Warning**: Deletion is permanent!

### Order Management

#### Viewing Orders
- Access through "🛒 سفارش‌ها"
- See order details:
  - Customer information
  - Ordered products
  - Order status
  - Payment status

#### Order Statuses
- **Pending**: New order, awaiting confirmation
- **Confirmed**: Order confirmed by you
- **Shipped**: Order shipped to customer
- **Delivered**: Order delivered successfully
- **Cancelled**: Order cancelled

### Sales Reports

#### Report Features
- Total sales amount
- Number of orders
- Commission calculations
- Date range filtering

#### Accessing Reports
1. Click "📈 گزارش فروش"
2. Select date range
3. View detailed statistics

### Store Settings

#### Available Settings
- **Welcome Message**: Custom greeting for customers
- **Support Contact**: Your contact information
- **Store Description**: Update store information

#### Customization Options (Pro/VIP)
- Custom button layouts
- Advanced welcome messages
- Premium advertisements

### Plan Renewal

#### Renewal Process
1. Click "🔄 تمدید پلن"
2. Select renewal duration:
   - 1 month
   - 3 months
   - 6 months
   - 12 months
3. Follow payment instructions
4. Send payment receipt
5. Wait for admin approval

#### Renewal Reminders
- 7 days before expiration
- 3 days before expiration
- Day of expiration

## Admin Guide

### Accessing Admin Panel
1. Send `/admin` command
2. Access admin features:
   - 📊 System statistics
   - 🏪 Store management
   - 💰 Payment management
   - 📢 Broadcast messages

### Store Management (Admin)

#### Pending Stores
- View new store registrations
- Approve or reject stores
- Activate store bots

#### Store Actions
- **✅ Approve**: Activate store and create sub-bot
- **❌ Reject**: Reject registration with reason
- **View Details**: Check store information

### Payment Management

#### Payment Verification
1. View pending payments
2. Check payment receipts
3. Approve or reject payments
4. Notify store owners

#### Payment Actions
- **✅ Approve**: Confirm payment and activate features
- **❌ Reject**: Reject payment with notification

### System Statistics

#### Available Metrics
- Total users count
- Active stores count
- Total revenue
- Monthly statistics

### Broadcasting Messages

#### Send Announcements
1. Click "📢 ارسال پیام همگانی"
2. Compose message
3. Select target audience:
   - All users
   - Store owners
   - Specific plan users
4. Send broadcast

## Sub-Bot Features

### Automatic Bot Creation
When a store is approved:
1. Unique bot is created automatically
2. Bot username: `storename_123_bot`
3. Bot token generated
4. Store owner receives bot credentials

### Sub-Bot Capabilities

#### Customer Features
- Browse products
- Add to cart
- Place orders
- Track order status
- Contact support

#### Advanced Features (Pro/VIP)
- Welcome messages
- Custom categories
- Advanced search
- Order history
- Payment tracking

### Bot Customization

#### Free Plan Limitations
- Standard buttons only
- Basic product display
- Limited customization

#### Pro Plan Features
- Custom welcome messages
- Enhanced product layouts
- Basic advertisements

#### VIP Plan Features
- Full customization
- Premium advertisements
- Custom button layouts
- Advanced analytics

## Support System

### Customer Support Options

#### FAQ Section
- Common questions and answers
- Step-by-step guides
- Troubleshooting tips

#### Contact Methods
- **❓ سوالات متداول**: Frequently asked questions
- **📞 تماس با ما**: Direct contact information
- **💬 پشتیبانی تلگرام**: Telegram support channel

### Response Times
- **Free Plan**: 24-48 hours
- **Pro Plan**: 12-24 hours (priority)
- **VIP Plan**: 2-6 hours (24/7 support)

## Best Practices

### Store Management
1. **Product Names**: Use clear, searchable names
2. **Descriptions**: Provide detailed product information
3. **Images**: Use high-quality product photos
4. **Pricing**: Keep prices competitive and updated
5. **Inventory**: Regularly update product availability

### Customer Service
1. **Response Time**: Reply to orders quickly
2. **Communication**: Keep customers informed
3. **Quality**: Ensure product quality matches description
4. **Support**: Provide helpful customer support

### Security
1. **Bot Token**: Never share your bot token
2. **Admin Access**: Protect admin credentials
3. **Payments**: Verify all payment receipts
4. **Data**: Keep customer information secure

## Troubleshooting

### Common Issues

#### Bot Not Responding
- Check if bot is active
- Verify internet connection
- Contact admin if persistent

#### Can't Add Products
- Check product limit for your plan
- Ensure all required fields are filled
- Verify image size (if uploading)

#### Payment Issues
- Ensure correct payment amount
- Use clear payment receipt photo
- Include transaction reference if available

#### Order Problems
- Check order status regularly
- Update inventory levels
- Communicate with customers

### Getting Help

#### Self-Help Resources
1. Check FAQ section
2. Review usage guide
3. Test with small orders first

#### Contacting Support
1. Use built-in support options
2. Provide clear problem description
3. Include relevant screenshots
4. Mention your plan type

#### Emergency Issues
For critical problems:
- Contact admin directly
- Use emergency contact methods
- Provide detailed error information

## Advanced Features

### API Integration (VIP)
- Custom payment gateways
- Third-party integrations
- Advanced analytics
- Webhook notifications

### Analytics and Reporting
- Customer behavior analysis
- Sales trend reports
- Product performance metrics
- Revenue optimization suggestions

### Multi-Store Management
- Manage multiple stores
- Cross-store promotions
- Centralized inventory
- Unified reporting

## Updates and Maintenance

### System Updates
- Regular feature updates
- Security patches
- Performance improvements
- New plan features

### Store Maintenance
- Regular product updates
- Inventory management
- Customer data cleanup
- Performance monitoring

For additional support and advanced features, contact: support@coderoot.ir