import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Pencil, PlusCircle, Trash2, Users } from 'lucide-react';
import { toast } from 'sonner';
import Navbar from '@/components/layout/Navbar';
import Container from '@/components/layout/Container';
import Footer from '@/components/layout/Footer';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui-custom/Card';
import Button from '@/components/ui-custom/Button';
import { Input } from '@/components/ui/input';
import { userService, authService } from '@/services/api';
import { getErrorMessage } from '@/lib/utils';

type Role = 'manager' | 'recruiter' | 'team_leader';

interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  tenantId?: string;
  companyName?: string;
  createdAt?: string;
}

interface Account {
  role: string;
  userCount: number;
  userLimit: number;
  companyName?: string;
}

const roleLabels: Record<Role, string> = {
  manager: 'Manager',
  recruiter: 'Recruiter',
  team_leader: 'Team Leader',
};

const emptyForm = { username: '', email: '', password: '', role: 'recruiter' as Role };

const AdminUsers = () => {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [form, setForm] = useState(emptyForm);

  const accountQuery = useQuery({
    queryKey: ['current-account'],
    queryFn: async () => (await authService.getCurrentAccount()).data.data as Account,
  });

  const usersQuery = useQuery({
    queryKey: ['admin-users'],
    queryFn: async () => (await userService.getAllUsers()).data.data as User[],
    enabled: accountQuery.data?.role === 'admin',
  });

  const createMutation = useMutation({
    mutationFn: () => userService.createUser(form),
    onSuccess: () => {
      toast.success('User created successfully');
      setShowCreate(false);
      setForm(emptyForm);
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      queryClient.invalidateQueries({ queryKey: ['current-account'] });
    },
    onError: (error) => toast.error(getErrorMessage(error, 'Could not create user.')),
  });

  const updateMutation = useMutation({
    mutationFn: () => {
      if (!editing) throw new Error('No user selected');
      return userService.updateUser(editing.id, {
        username: form.username,
        email: form.email,
        role: form.role,
      });
    },
    onSuccess: () => {
      toast.success('User updated successfully');
      setEditing(null);
      setForm(emptyForm);
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
    },
    onError: (error) => toast.error(getErrorMessage(error, 'Could not update user.')),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => userService.deleteUser(id),
    onSuccess: () => {
      toast.success('User deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      queryClient.invalidateQueries({ queryKey: ['current-account'] });
    },
    onError: (error) => toast.error(getErrorMessage(error, 'Could not delete user.')),
  });

  const openCreate = () => {
    setEditing(null);
    setForm(emptyForm);
    setShowCreate(true);
  };

  const openEdit = (user: User) => {
    setShowCreate(false);
    setEditing(user);
    setForm({
      username: user.username,
      email: user.email,
      password: '',
      role: (user.role === 'manager' || user.role === 'team_leader' ? user.role : 'recruiter'),
    });
  };

  const submit = (event: React.FormEvent) => {
    event.preventDefault();
    if (editing) {
      updateMutation.mutate();
    } else {
      createMutation.mutate();
    }
  };

  const isSaving = createMutation.isPending || updateMutation.isPending;

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar />
      <main className="pt-24 pb-10 flex-grow">
        <Container>
          <div className="mb-8">
            <h1 className="text-3xl font-semibold tracking-tight mb-2">User Management</h1>
            <p className="text-ats-gray-500">Manage users and roles for this SkillSifter tenant.</p>
          </div>

          {accountQuery.isLoading ? (
            <Card><CardContent className="p-8 text-center">Loading account...</CardContent></Card>
          ) : accountQuery.data?.role !== 'admin' ? (
            <Card><CardContent className="p-8 text-center text-red-600">Administrator access is required.</CardContent></Card>
          ) : (
            <>
              <Card className="mb-6">
                <CardContent className="p-6 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <Users className="w-6 h-6 text-ats-blue-500" />
                    <div>
                      <div className="font-semibold">Seats</div>
                      <div className="text-sm text-ats-gray-500">
                        {accountQuery.data.userCount} of {accountQuery.data.userLimit} users in use
                      </div>
                    </div>
                  </div>
                  <Button
                    variant="primary"
                    size="sm"
                    className="flex gap-2"
                    onClick={openCreate}
                    disabled={accountQuery.data.userCount >= accountQuery.data.userLimit}
                  >
                    <PlusCircle size={16} />
                    Add User
                  </Button>
                </CardContent>
              </Card>

              {(showCreate || editing) && (
                <Card className="mb-6">
                  <CardHeader>
                    <CardTitle>{editing ? 'Edit User' : 'Add User'}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <form onSubmit={submit} className="grid gap-4 md:grid-cols-2">
                      <div>
                        <label className="text-sm font-medium">Name</label>
                        <Input
                          value={form.username}
                          onChange={(e) => setForm({ ...form, username: e.target.value })}
                          placeholder="User name"
                          required
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium">Email</label>
                        <Input
                          type="email"
                          value={form.email}
                          onChange={(e) => setForm({ ...form, email: e.target.value })}
                          placeholder="user@example.com"
                          required
                        />
                      </div>
                      {!editing && (
                        <div>
                          <label className="text-sm font-medium">Password</label>
                          <Input
                            type="password"
                            value={form.password}
                            onChange={(e) => setForm({ ...form, password: e.target.value })}
                            placeholder="Initial password"
                            required
                          />
                        </div>
                      )}
                      <div>
                        <label className="text-sm font-medium">Role</label>
                        <select
                          className="mt-1 flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                          value={form.role}
                          onChange={(e) => setForm({ ...form, role: e.target.value as Role })}
                        >
                          {Object.entries(roleLabels).map(([value, label]) => (
                            <option key={value} value={value}>{label}</option>
                          ))}
                        </select>
                      </div>
                      <div className="md:col-span-2 flex gap-2">
                        <Button type="submit" variant="primary" disabled={isSaving}>
                          {isSaving ? 'Saving...' : editing ? 'Save Changes' : 'Create User'}
                        </Button>
                        <Button
                          type="button"
                          variant="outline"
                          onClick={() => { setShowCreate(false); setEditing(null); }}
                        >
                          Cancel
                        </Button>
                      </div>
                    </form>
                  </CardContent>
                </Card>
              )}

              <Card>
                <CardHeader>
                  <CardTitle>Tenant Users</CardTitle>
                </CardHeader>
                <CardContent>
                  {usersQuery.isLoading ? (
                    <div className="p-8 text-center text-ats-gray-500">Loading users...</div>
                  ) : usersQuery.isError ? (
                    <div className="p-8 text-center text-red-600">Could not load users.</div>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b text-left">
                            <th className="p-3">Name</th>
                            <th className="p-3">Email</th>
                            <th className="p-3">Role</th>
                            <th className="p-3 text-right">Actions</th>
                          </tr>
                        </thead>
                        <tbody>
                          {(usersQuery.data ?? []).map((user) => {
                            const isAdmin = user.role === 'admin';
                            return (
                              <tr key={user.id} className="border-b last:border-0">
                                <td className="p-3 font-medium">{user.username}</td>
                                <td className="p-3">{user.email}</td>
                                <td className="p-3 capitalize">{user.role.replace('_', ' ')}</td>
                                <td className="p-3 text-right">
                                  {isAdmin ? (
                                    <span className="text-xs text-ats-gray-500">Tenant administrator</span>
                                  ) : (
                                    <div className="flex justify-end gap-2">
                                      <Button type="button" variant="outline" size="sm" onClick={() => openEdit(user)}>
                                        <Pencil size={15} />
                                      </Button>
                                      <Button
                                        type="button"
                                        variant="outline"
                                        size="sm"
                                        onClick={() => {
                                          if (window.confirm(`Delete user ${user.username}?`)) {
                                            deleteMutation.mutate(user.id);
                                          }
                                        }}
                                      >
                                        <Trash2 size={15} />
                                      </Button>
                                    </div>
                                  )}
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    </div>
                  )}
                </CardContent>
              </Card>
            </>
          )}
        </Container>
      </main>
      <Footer />
    </div>
  );
};

export default AdminUsers;
