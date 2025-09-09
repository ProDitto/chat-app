import { useState } from 'react';
import { Button } from '../components/ui/Button';
import { useAuthStore } from '../store/authStore';
import { toast } from '../hooks/use-toast';
// import { Input } from '../components/ui/Input';
import UserAvatar from '../components/common/UserAvatar';
import { LayoutDashboard, LogOut, User, ImageUp } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useTheme } from '../components/theme/ThemeProvider.tsx';
import { Switch } from '../components/ui/Switch.tsx'; // Ensure you have this component
import { uploadProfilePicture } from '../api/user.ts';

const SettingsPage = () => {
  const { user, setUser } = useAuthStore();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const { theme, setTheme } = useTheme();

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setSelectedFile(event.target.files[0]);
    } else {
      setSelectedFile(null);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile || !user) {
      toast({ title: 'Error', description: 'No file selected or user not logged in.', variant: 'destructive' });
      return;
    }
    if (selectedFile.size > 200 * 1024) { // 200 KB limit
        toast({ title: 'Error', description: 'Image size limit is 200KB.', variant: 'destructive' });
        return;
    }

    setIsUploading(true);
    try {
      const formData = new FormData();
      formData.append('profile_picture', selectedFile);
      const newUrl = await uploadProfilePicture(formData);
      setUser({ ...user, profilePictureUrl: newUrl });
      toast({ title: 'Success', description: 'Profile picture updated!' });
      setSelectedFile(null);
    } catch (error: any) {
      toast({ title: 'Error', description: error.response?.data?.error || 'Failed to upload profile picture.', variant: 'destructive' });
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="flex h-screen bg-background-secondary">
      {/* Sidebar for settings navigation */}
      <aside className="w-64 border-r border-border bg-background-primary p-4 flex flex-col">
        <h2 className="text-2xl font-bold font-heading mb-6">Settings</h2>
        <nav className="space-y-2 flex-1">
          <Link to="/" className="flex items-center p-2 rounded-md hover:bg-background-secondary text-text-primary">
            <LayoutDashboard className="mr-2 h-4 w-4" /> Dashboard
          </Link>
          <div className="flex items-center p-2 rounded-md bg-primary-accent/10 text-primary-accent font-semibold">
            <User className="mr-2 h-4 w-4" /> Profile
          </div>
          {/* Add other setting categories here */}
        </nav>
        <div className="mt-auto pt-4 border-t border-border">
            <div className="flex items-center justify-between p-2">
                <span>Dark Mode</span>
                <Switch 
                  checked={theme === 'dark'} 
                  onCheckedChange={(checked) => setTheme(checked ? 'dark' : 'light')} 
                />
            </div>
            <Button variant="ghost" className="w-full justify-start text-status-error" onClick={useAuthStore.getState().logout}>
                <LogOut className="mr-2 h-4 w-4" /> Logout
            </Button>
        </div>
      </aside>

      {/* Main content area for profile settings */}
      <main className="flex-1 p-8 overflow-y-auto">
        <h1 className="text-3xl font-bold font-heading mb-8">Profile Settings</h1>

        <section className="bg-background-primary p-6 rounded-lg shadow-md mb-8">
          <h2 className="text-xl font-semibold mb-4">Profile Picture</h2>
          <div className="flex items-center space-x-4">
            <UserAvatar user={user || {}} size="lg" />
            <div>
              <input type="file" accept="image/*" onChange={handleFileChange} className="block w-full text-sm text-text-secondary file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-primary-accent file:text-white hover:file:bg-primary-accent/90" />
              {selectedFile && <p className="text-sm text-text-secondary mt-2">Selected: {selectedFile.name}</p>}
              <Button onClick={handleUpload} disabled={!selectedFile || isUploading} className="mt-4">
                <ImageUp className="mr-2 h-4 w-4" /> {isUploading ? 'Uploading...' : 'Upload Picture'}
              </Button>
            </div>
          </div>
        </section>

        {/* Add other profile settings sections here (e.g., change username, password) */}
      </main>
    </div>
  );
};

export default SettingsPage;
