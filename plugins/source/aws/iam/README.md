aws iam plugins
===============

# aws_iam_users

#### configuration

- `omit_user_tags` (bool): when true, do not include the user's resource tags as metric tags
- `omit_user_path` (bool): when true, do not include the user's path as a metric tag

#### output

Produce one datum for each IAM user found in the given session.

**name:** `aws_iam_user`
**tags:**

- `user_name`: the user's name
- `user_path`: the user's iam resource path (unless `omit_user_path` is true)
- `tag_{tagname}`: one field for each iam resource tag on the user (unless `omit_user_tags` is true)

**fields:**

- `age` (duration): the length of the since the user was created
- `active_key_count` (count): the number of active access keys associated with the user
- `oldest_key_age` (duration): the length of time since the oldest access key was created
- `since_last_activity` (duration): the length of time since any activity by the user
- `since_last_login_activity` (duration): the length of time since the user last logged in with their password
- `since_last_key_activity` (duration): the length of time since the user last used an access key
