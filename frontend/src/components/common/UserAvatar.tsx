import type { User } from '../../types/user';

interface UserAvatarProps {
  user: Partial<User>;
  size?: 'sm' | 'md' | 'lg';
}

const sizeClasses = {
  sm: 'w-8 h-8 text-sm',
  md: 'w-10 h-10 text-base',
  lg: 'w-12 h-12 text-lg',
};

const getInitials = (name: string) => {
  return name
    .split(' ')
    .map((n) => n[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();
};

const UserAvatar = ({ user, size = 'md' }: UserAvatarProps) => {
  return (
    <div
      className={`relative inline-flex items-center justify-center overflow-hidden bg-gray-600 rounded-full ${sizeClasses[size]}`}
    >
      {user.profilePictureUrl ? (
        <img className="object-cover w-full h-full" src={user.profilePictureUrl} alt={user.username} />
      ) : (
        <span className="font-medium text-gray-300">
          {user.username ? getInitials(user.username) : '??'}
        </span>
      )}
    </div>
  );
};

export default UserAvatar;
