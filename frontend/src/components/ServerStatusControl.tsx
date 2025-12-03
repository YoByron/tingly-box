import { Cancel, CheckCircle } from '@mui/icons-material';
import {
    Button,
    Stack,
    Typography,
} from '@mui/material';

interface ServerStatusControlProps {
    serverStatus: any;
    onStartServer: () => void;
    onStopServer: () => void;
    onRestartServer: () => void;
    onGenerateToken?: () => void;
}

const ServerStatusControl = ({
    serverStatus,
    onStartServer,
    onStopServer,
    onRestartServer,
    onGenerateToken,
}: ServerStatusControlProps) => {
    return (
        <Stack spacing={2}>
            <Stack direction="row" alignItems="center" spacing={1}>
                {serverStatus.server_running ? (
                    <CheckCircle color="success" />
                ) : (
                    <Cancel color="error" />
                )}
                <Typography variant="h6">
                    Status: {serverStatus.server_running ? 'Running' : 'Stopped'}
                </Typography>
            </Stack>
            <Typography variant="body2">
                <strong>Port:</strong> {serverStatus.port}
            </Typography>
            <Typography variant="body2">
                <strong>Providers:</strong> {serverStatus.providers_enabled}/{serverStatus.providers_total}
            </Typography>
            {serverStatus.uptime && (
                <Typography variant="body2">
                    <strong>Uptime:</strong> {serverStatus.uptime}
                </Typography>
            )}
            {serverStatus.last_updated && (
                <Typography variant="body2">
                    <strong>Last Updated:</strong> {serverStatus.last_updated}
                </Typography>
            )}
            {serverStatus.request_count !== undefined && (
                <Typography variant="body2">
                    <strong>Request Count:</strong> {serverStatus.request_count}
                </Typography>
            )}
            <Stack direction="row" spacing={2} sx={{ mt: 3 }}>
                <Button
                    variant="contained"
                    color="success"
                    onClick={onStartServer}
                    disabled={serverStatus.server_running}
                >
                    Start
                </Button>
                <Button
                    variant="contained"
                    color="error"
                    onClick={onStopServer}
                    disabled={!serverStatus.server_running}
                >
                    Stop
                </Button>
                <Button variant="contained" onClick={onRestartServer}>
                    Restart
                </Button>
                {onGenerateToken && (
                    <Button variant="outlined" onClick={onGenerateToken}>
                        Generate Token
                    </Button>
                )}
            </Stack>
        </Stack>
    );
};

export default ServerStatusControl;