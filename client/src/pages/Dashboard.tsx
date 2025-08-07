import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Store, Package, ShoppingCart, Users, TrendingUp, Bot, TestTube } from "lucide-react";
import { Link } from "wouter";

const Dashboard = () => {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-bold mb-2">Telegram Store Hub</h1>
              <p className="text-muted-foreground text-lg">
                Multi-store e-commerce bot management platform
              </p>
            </div>
            <Link href="/api-test">
              <Button variant="outline" className="gap-2" data-testid="link-api-test">
                <TestTube className="h-4 w-4" />
                API Testing
              </Button>
            </Link>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card data-testid="card-active-stores">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Stores</CardTitle>
              <Store className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold" data-testid="text-active-stores-count">12</div>
              <p className="text-xs text-muted-foreground" data-testid="text-active-stores-change">+2 from last month</p>
            </CardContent>
          </Card>

          <Card data-testid="card-total-products">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Products</CardTitle>
              <Package className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold" data-testid="text-total-products-count">1,234</div>
              <p className="text-xs text-muted-foreground" data-testid="text-total-products-change">+15% from last month</p>
            </CardContent>
          </Card>

          <Card data-testid="card-orders-today">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Orders Today</CardTitle>
              <ShoppingCart className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold" data-testid="text-orders-today-count">89</div>
              <p className="text-xs text-muted-foreground" data-testid="text-orders-today-change">+7% from yesterday</p>
            </CardContent>
          </Card>

          <Card data-testid="card-active-users">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Users</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold" data-testid="text-active-users-count">2,456</div>
              <p className="text-xs text-muted-foreground" data-testid="text-active-users-change">+12% from last week</p>
            </CardContent>
          </Card>
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Bot Status */}
          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Bot className="h-5 w-5" />
                Bot Status & Performance
              </CardTitle>
              <CardDescription>
                Monitor your Telegram bots across all stores
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <h4 className="font-medium">Main Store Bot</h4>
                    <p className="text-sm text-muted-foreground">@mainstore_bot</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="bg-green-100 text-green-800">
                      Online
                    </Badge>
                    <span className="text-sm text-muted-foreground">99.9% uptime</span>
                  </div>
                </div>
                
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <h4 className="font-medium">Electronics Store Bot</h4>
                    <p className="text-sm text-muted-foreground">@electronics_bot</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="bg-green-100 text-green-800">
                      Online
                    </Badge>
                    <span className="text-sm text-muted-foreground">98.7% uptime</span>
                  </div>
                </div>

                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <h4 className="font-medium">Fashion Store Bot</h4>
                    <p className="text-sm text-muted-foreground">@fashion_bot</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="bg-yellow-100 text-yellow-800">
                      Maintenance
                    </Badge>
                    <span className="text-sm text-muted-foreground">97.2% uptime</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
              <CardDescription>
                Common management tasks
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <Button className="w-full justify-start" variant="outline" data-testid="button-create-store">
                <Store className="h-4 w-4 mr-2" />
                Create New Store
              </Button>
              
              <Button className="w-full justify-start" variant="outline" data-testid="button-add-products">
                <Package className="h-4 w-4 mr-2" />
                Add Products
              </Button>
              
              <Button className="w-full justify-start" variant="outline" data-testid="button-view-analytics">
                <TrendingUp className="h-4 w-4 mr-2" />
                View Analytics
              </Button>
              
              <Button className="w-full justify-start" variant="outline" data-testid="button-bot-settings">
                <Bot className="h-4 w-4 mr-2" />
                Bot Settings
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Recent Activity */}
        <Card className="mt-6">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>
              Latest orders and system events
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 border rounded-lg">
                <div>
                  <h4 className="font-medium">New order #1234</h4>
                  <p className="text-sm text-muted-foreground">Electronics Store • 2 items • $299.99</p>
                </div>
                <Badge>Pending</Badge>
              </div>
              
              <div className="flex items-center justify-between p-4 border rounded-lg">
                <div>
                  <h4 className="font-medium">Product "iPhone 15" updated</h4>
                  <p className="text-sm text-muted-foreground">Main Store • Price changed to $899.99</p>
                </div>
                <Badge variant="secondary">Updated</Badge>
              </div>
              
              <div className="flex items-center justify-between p-4 border rounded-lg">
                <div>
                  <h4 className="font-medium">New customer registered</h4>
                  <p className="text-sm text-muted-foreground">Fashion Store • @username_123</p>
                </div>
                <Badge variant="outline">New User</Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Dashboard;