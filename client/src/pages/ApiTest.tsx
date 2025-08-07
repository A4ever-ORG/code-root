import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/hooks/use-toast";
import { apiRequest } from "@/lib/queryClient";
import { LoadingSpinner } from "@/components/LoadingSpinner";
import { Code, Send, Users, Shield, Home } from "lucide-react";
import { Link } from "wouter";

const ApiTest = () => {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<string>("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const testEndpoint = async (method: string, endpoint: string, data?: any) => {
    setLoading(true);
    try {
      let result;
      if (method === "GET") {
        const response = await fetch(endpoint);
        result = await response.json();
      } else {
        result = await apiRequest(endpoint, {
          method,
          body: JSON.stringify(data),
        });
      }
      
      setResponse(JSON.stringify(result, null, 2));
      toast({
        title: "Request Successful",
        description: `${method} ${endpoint}`,
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error";
      setResponse(`Error: ${errorMessage}`);
      toast({
        title: "Request Failed",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const createUser = () => {
    if (!username || !password) {
      toast({
        title: "Validation Error",
        description: "Username and password are required",
        variant: "destructive",
      });
      return;
    }
    testEndpoint("POST", "/api/users", { username, password });
  };

  const testSecurity = () => {
    // Test XSS prevention
    testEndpoint("POST", "/api/users", { 
      username: "test<script>alert('xss')</script>", 
      password: "testpass123" 
    });
  };

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-bold mb-2">API Testing Dashboard</h1>
              <p className="text-muted-foreground text-lg">
                Test and validate API endpoints with comprehensive security checks
              </p>
            </div>
            <Link href="/">
              <Button variant="outline" className="gap-2" data-testid="link-home">
                <Home className="h-4 w-4" />
                Back to Dashboard
              </Button>
            </Link>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* API Testing */}
          <Card data-testid="card-api-testing">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Send className="h-5 w-5" />
                API Endpoints
              </CardTitle>
              <CardDescription>
                Test various API endpoints and operations
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <Button 
                  onClick={() => testEndpoint("GET", "/api/health")}
                  variant="outline"
                  disabled={loading}
                  data-testid="button-test-health"
                >
                  Health Check
                </Button>
                <Button 
                  onClick={() => testEndpoint("GET", "/api/users")}
                  variant="outline"
                  disabled={loading}
                  data-testid="button-test-users"
                >
                  Get Users
                </Button>
              </div>
              
              <Separator />
              
              <div className="space-y-3">
                <div>
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    placeholder="Enter username"
                    data-testid="input-username"
                  />
                </div>
                <div>
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter password"
                    data-testid="input-password"
                  />
                </div>
                <Button 
                  onClick={createUser}
                  disabled={loading}
                  className="w-full"
                  data-testid="button-create-user"
                >
                  {loading ? <LoadingSpinner size="sm" /> : <Users className="h-4 w-4 mr-2" />}
                  Create User
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Security Testing */}
          <Card data-testid="card-security-testing">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Security Tests
              </CardTitle>
              <CardDescription>
                Validate security measures and input sanitization
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3">
                <Button 
                  onClick={testSecurity}
                  variant="outline"
                  disabled={loading}
                  className="w-full"
                  data-testid="button-test-xss"
                >
                  Test XSS Protection
                </Button>
                
                <Button 
                  onClick={() => testEndpoint("GET", "/api/users/999")}
                  variant="outline"
                  disabled={loading}
                  className="w-full"
                  data-testid="button-test-404"
                >
                  Test 404 Handling
                </Button>
                
                <Button 
                  onClick={() => testEndpoint("GET", "/api/users/abc")}
                  variant="outline"
                  disabled={loading}
                  className="w-full"
                  data-testid="button-test-invalid-id"
                >
                  Test Invalid ID
                </Button>
              </div>
              
              <div className="flex flex-wrap gap-2 pt-3">
                <Badge variant="secondary">XSS Protection</Badge>
                <Badge variant="secondary">Input Validation</Badge>
                <Badge variant="secondary">Error Handling</Badge>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Response Display */}
        <Card data-testid="card-response">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Code className="h-5 w-5" />
              API Response
            </CardTitle>
            <CardDescription>
              Latest API response data
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Textarea
              value={response}
              readOnly
              placeholder="API responses will appear here..."
              className="min-h-[200px] font-mono text-sm"
              data-testid="textarea-response"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default ApiTest;