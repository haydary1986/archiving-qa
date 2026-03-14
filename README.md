# نظام الأرشفة الإلكتروني - قسم ضمان الجودة وتقييم الأداء

## نظرة عامة

نظام أرشفة إلكتروني متكامل مصمم لقسم ضمان الجودة وتقييم الأداء في الجامعة. يوفر النظام إدارة شاملة للوثائق والكتب الرسمية مع دعم الذكاء الاصطناعي والتكامل مع Google Workspace.

## البنية التقنية

| المكون | التقنية |
|--------|---------|
| الواجهة الأمامية | Nuxt.js 3 + Nuxt UI + TailwindCSS |
| الواجهة الخلفية | Go (Gin Framework) |
| قاعدة البيانات | PostgreSQL 16 |
| التخزين المؤقت | Redis 7 |
| طوابير المهام | Asynq (Redis-based) |
| تخزين الملفات | Google Drive API |
| OCR | Tesseract OCR |
| ذكاء اصطناعي | Ollama/Gemini/DeepSeek |
| النشر | Docker + Coolify |

## الميزات الرئيسية

### إدارة الوثائق
- إنشاء وتعديل وحذف الوثائق (حذف منطقي)
- تصنيف الوثائق (وارد/صادر/داخلي)
- حقول مخصصة ديناميكية (JSONB)
- بحث متقدم بالعنوان والرقم والنص الكامل
- تصفية حسب النوع والتصنيف والحالة والتاريخ

### الملفات والتخزين
- رفع الملفات (PDF, Word, Images) إلى Google Drive
- ضغط تلقائي للصور وملفات PDF
- متوافق مع الهاتف لالتقاط الصور
- روابط مشاركة مؤمنة بكلمة مرور

### المعالجة الذكية
- **OCR**: استخراج النص من الصور والملفات عبر Tesseract
- **AI**: تحليل الوثائق واستخراج البيانات تلقائياً (عنوان الكتاب، الجهة، الرقم، التاريخ)
- دعم Ollama (محلي)، Gemini، DeepSeek

### تتبع الوثائق
- ربط الأشخاص بالوثائق (مرسل/مستلم/ذو علاقة)
- تتبع مسار الوارد والصادر (Routing)
- سجل كامل لتحويلات الكتب بين الجهات

### الأمان والصلاحيات
- **4 أدوار رئيسية:**
  - `super_admin` - مدير النظام: صلاحيات كاملة
  - `qa_manager` - مدير الجودة: إدارة الوثائق والمستخدمين
  - `data_entry` - مدخل بيانات: إضافة وتعديل الوثائق
  - `viewer` - مشاهد: قراءة فقط (قابل للتقييد بأقسام محددة)
- تسجيل دخول Google SSO + محلي
- سجل تدقيق شامل (Audit Log)
- تقييد الوصول حسب الأقسام

### التصدير والتقارير
- تصدير البيانات إلى Excel
- تجميع الملفات في ZIP

## التشغيل السريع

### المتطلبات
- Docker و Docker Compose
- حساب Google Cloud (لـ Drive API)

### الخطوات

```bash
# 1. استنساخ المشروع
git clone https://github.com/haydary1986/archiving-qa.git
cd archiving-qa

# 2. إعداد ملف البيئة
cp .env.example .env
# عدّل ملف .env بالقيم المناسبة

# 3. تشغيل النظام
docker compose up -d

# 4. (اختياري) تشغيل مع AI المحلي
docker compose --profile ai up -d

# 5. (اختياري) تشغيل مع محرك البحث
docker compose --profile search up -d
```

### الوصول
- **الواجهة الأمامية**: http://localhost:3000
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### حساب المدير الافتراضي
- البريد: `admin@university.edu.iq`
- كلمة المرور: `Admin@123456`
- **يجب تغيير كلمة المرور فور تسجيل الدخول**

## هيكل المشروع

```
archiving-qa/
├── backend/
│   ├── cmd/server/          # نقطة الدخول الرئيسية
│   ├── internal/
│   │   ├── config/          # إعدادات النظام
│   │   ├── database/        # اتصال قاعدة البيانات والـ migrations
│   │   ├── models/          # نماذج البيانات
│   │   ├── handlers/        # معالجات API
│   │   ├── middleware/      # وسيط المصادقة والصلاحيات
│   │   ├── services/        # خدمات (Drive, OCR, AI, Compress, Export)
│   │   ├── routes/          # تعريف المسارات
│   │   └── workers/         # مهام الخلفية (OCR, AI)
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── assets/css/          # أنماط CSS
│   ├── components/          # مكونات Vue
│   ├── composables/         # دوال مشتركة (useApi)
│   ├── layouts/             # تخطيطات الصفحات
│   ├── middleware/           # وسيط التوجيه
│   ├── pages/               # صفحات التطبيق
│   ├── plugins/             # إضافات
│   ├── stores/              # مخازن Pinia
│   ├── types/               # أنواع TypeScript
│   ├── Dockerfile
│   └── nuxt.config.ts
├── docker-compose.yml
├── .env.example
└── README.md
```

## واجهات API

### المصادقة
| Method | Endpoint | الوصف |
|--------|----------|-------|
| POST | `/api/v1/auth/login` | تسجيل الدخول |
| POST | `/api/v1/auth/register` | إنشاء حساب |
| POST | `/api/v1/auth/refresh` | تحديث الرمز |
| GET | `/api/v1/auth/profile` | الملف الشخصي |
| GET | `/api/v1/auth/google/callback` | Google OAuth |

### الوثائق
| Method | Endpoint | الوصف |
|--------|----------|-------|
| GET | `/api/v1/documents` | قائمة الوثائق (مع فلترة وصفحات) |
| GET | `/api/v1/documents/:id` | تفاصيل وثيقة |
| POST | `/api/v1/documents` | إنشاء وثيقة |
| PUT | `/api/v1/documents/:id` | تعديل وثيقة |
| DELETE | `/api/v1/documents/:id` | حذف منطقي |
| POST | `/api/v1/documents/:id/restore` | استعادة من المحذوفات |
| POST | `/api/v1/documents/:id/files` | رفع ملف |

### الأشخاص
| Method | Endpoint | الوصف |
|--------|----------|-------|
| GET | `/api/v1/persons` | قائمة الأشخاص |
| GET | `/api/v1/persons/:id` | تفاصيل شخص + وثائقه |
| POST | `/api/v1/persons` | إضافة شخص |
| PUT | `/api/v1/persons/:id` | تعديل |
| DELETE | `/api/v1/persons/:id` | حذف |

### التصنيفات والوسوم
| Method | Endpoint | الوصف |
|--------|----------|-------|
| GET | `/api/v1/categories` | شجرة التصنيفات |
| POST | `/api/v1/categories` | إضافة تصنيف |
| GET | `/api/v1/tags` | قائمة الوسوم |
| POST | `/api/v1/tags` | إضافة وسم |

### الإدارة (Admin)
| Method | Endpoint | الوصف |
|--------|----------|-------|
| GET | `/api/v1/admin/users` | إدارة المستخدمين |
| GET | `/api/v1/admin/roles` | إدارة الأدوار |
| GET | `/api/v1/admin/permissions` | قائمة الصلاحيات |
| GET | `/api/v1/admin/audit-logs` | سجل التدقيق |
| GET | `/api/v1/admin/settings` | إعدادات النظام |
| GET | `/api/v1/admin/trash` | سلة المحذوفات |

### أخرى
| Method | Endpoint | الوصف |
|--------|----------|-------|
| POST | `/api/v1/share` | إنشاء رابط مشاركة |
| GET | `/api/v1/share/:token` | الوصول عبر رابط مشاركة |
| POST | `/api/v1/export` | تصدير وثائق |
| GET | `/api/v1/dashboard` | إحصائيات لوحة التحكم |
| POST | `/api/v1/routings` | إضافة مسار توجيه |

## النشر عبر Coolify

1. أنشئ مشروع جديد في Coolify
2. اربطه بمستودع GitHub: `https://github.com/haydary1986/archiving-qa`
3. اختر Docker Compose كنوع النشر
4. أضف متغيرات البيئة من ملف `.env.example`
5. انشر المشروع

## قاعدة البيانات

### الجداول الرئيسية
- `users` - المستخدمون
- `roles` - الأدوار
- `permissions` - الصلاحيات
- `role_permissions` - ربط الأدوار بالصلاحيات
- `user_category_access` - تقييد وصول المشاهدين لأقسام محددة
- `documents` - الوثائق
- `files` - الملفات المرفقة
- `categories` - التصنيفات (شجرية)
- `tags` - الوسوم
- `persons` - الأشخاص
- `document_persons` - ربط الوثائق بالأشخاص
- `routings` - مسارات التوجيه
- `audit_logs` - سجل التدقيق (غير قابل للتعديل)
- `custom_field_defs` - تعريفات الحقول المخصصة
- `system_settings` - إعدادات النظام
- `share_links` - روابط المشاركة

## الترخيص

خاص - جامعة (اسم الجامعة) - قسم ضمان الجودة وتقييم الأداء
