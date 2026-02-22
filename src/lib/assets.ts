export interface SteamAsset {
    id: string;
    name: string;
    width: number;
    height: number;
}

export const STEAM_ASSETS: SteamAsset[] = [
    { id: 'header-capsule',         name: 'Header Capsule',         width: 920,  height: 430  },
    { id: 'small-capsule',          name: 'Small Capsule',          width: 462,  height: 174  },
    { id: 'main-capsule',           name: 'Main Capsule',           width: 1232, height: 706  },
    { id: 'vertical-capsule',       name: 'Vertical Capsule',       width: 748,  height: 896  },
    { id: 'screenshots',            name: 'Screenshots',            width: 1920, height: 1080 },
    { id: 'page-background',        name: 'Page Background',        width: 1438, height: 810  },
    { id: 'shortcut-icon',          name: 'Shortcut Icon',          width: 256,  height: 256  },
    { id: 'app-icon',               name: 'App Icon',               width: 184,  height: 184  },
    { id: 'library-capsule',        name: 'Library Capsule',        width: 600,  height: 900  },
    { id: 'library-hero',           name: 'Library Hero',           width: 3840, height: 1240 },
    { id: 'library-logo',           name: 'Library Logo',           width: 1280, height: 720  },
    { id: 'library-header-capsule', name: 'Library Header Capsule', width: 920,  height: 430  },
    { id: 'event-cover',            name: 'Event Cover',            width: 800,  height: 450  },
    { id: 'event-header',           name: 'Event Header',           width: 1920, height: 622  },
];
