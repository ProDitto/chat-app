import { api } from './client';

export const uploadProfilePicture = async (formData: FormData): Promise<string> => {
  const response = await api.post('/me/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data.profile_picture_url;
};