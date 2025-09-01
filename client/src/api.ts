import { error, log } from './utils';

export type Guild = {
    id: string;
    name: string;
    icon: string;
};

export type Category = {
    id?: number;
    name: string;
    guildId: string;
    sort?: number;
    created?: Date;
    updated?: Date;
};

export type Sound = {
    id?: number;
    name: string;
    hash: string;
    categoryId: number;
    sort?: number;
    created?: Date;
    updated?: Date;
};

export const api = {
    async listGuilds(): Promise<Guild[]> {
        log('Fetch user guild list');
        const resp = await fetch('/api/list_user_guilds');
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
        const { guilds } = await resp.json();
        return guilds;
    },

    async listCategories(guildId: string): Promise<Category[]> {
        log(`Fetch categories for guild ${guildId}`);
        const resp = await fetch('/api/list_guild_categories?guild=' + guildId);
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
        const { categories } = await resp.json();
        return categories;
    },

    async listSounds(guildId: string): Promise<Sound[]> {
        log(`Fetch sounds for guild ${guildId}`);
        const resp = await fetch('/api/list_guild_sounds?guild=' + guildId);
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
        const { sounds } = await resp.json();
        return sounds;
    },

    async uploadSound(data: ArrayBuffer): Promise<string> {
        log(`Upload sound`);
        const resp = await fetch('/api/upload_sound', {
            method: 'POST',
            body: data,
        });
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
        const { hash } = await resp.json();
        return hash;
    },

    async saveCategory(category: Category): Promise<Category> {
        log(`Save category ${category.id ?? '(new)'}`);
        const resp = await fetch('/api/category', {
            method: category.id ? 'PUT' : 'POST',
            body: JSON.stringify(category),
        });
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
        const { category: updateCategory } = await resp.json();
        return updateCategory;
    },

    async deleteCategory(category: Category): Promise<void> {
        log(`Delete category ${category.id}`);
        const resp = await fetch('/api/category', {
            method: 'DELETE',
            body: JSON.stringify({ id: category.id }),
        });
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
    },

    async saveSound(sound: Sound): Promise<Sound> {
        log(`Save sound with hash ${sound.hash}`);
        const resp = await fetch('/api/sound', {
            method: sound.id ? 'PUT' : 'POST',
            body: JSON.stringify(sound),
        });
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
        const { sound: updateSound } = await resp.json();
        return updateSound;
    },

    async deleteSound(sound: Sound): Promise<void> {
        log(`Delete sound ${sound.id}`);
        const resp = await fetch('/api/sound', {
            method: 'DELETE',
            body: JSON.stringify({ id: sound.id }),
        });
        if (!resp.ok) {
            const { error: errorMsg } = await resp.json();
            error(errorMsg);
            throw new Error(errorMsg);
        }
    },
};
